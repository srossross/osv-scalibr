// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package uvlock extracts uv.lock files.
package uvlock

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/internal/depgraph"
	"github.com/google/osv-scalibr/extractor/filesystem/language/python/internal/pep503"
	"github.com/google/osv-scalibr/extractor/filesystem/language/python/internal/tomldep"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/plugin"
	"github.com/google/osv-scalibr/purl"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

const (
	// Name is the unique name of this extractor.
	Name = "python/uvlock"
)

type uvLockPackageSource struct {
	Virtual  string `toml:"virtual"`
	Editable string `toml:"editable"`
	Git      string `toml:"git"`
}

// isScannedProject reports whether the package is the project being scanned
// rather than one of its dependencies. uv writes "virtual" for a project
// without a build system and "editable" for a packaged one.
func (s uvLockPackageSource) isScannedProject() bool {
	return s.Virtual == "." || s.Editable == "."
}

type uvLockPackage struct {
	Name    string              `toml:"name"`
	Version string              `toml:"version"`
	Source  uvLockPackageSource `toml:"source"`

	Dependencies []tomldep.Dependency `toml:"dependencies"`

	// uv stores "groups" as a table under "package" after all the packages, which due
	// to how TOML works means it ends up being a property on the last package, even
	// through in this context it's a global property rather than being per-package
	Groups map[string][]tomldep.Dependency `toml:"optional-dependencies"`

	DevGroups map[string][]tomldep.Dependency `toml:"dev-dependencies"`
}

type uvLockFile struct {
	Version  int             `toml:"version"`
	Packages []uvLockPackage `toml:"package"`
}

// Extractor extracts python packages from uv.lock files.
type Extractor struct{}

// New returns a new instance of the extractor.
func New(_ *cpb.PluginConfig) (filesystem.Extractor, error) { return &Extractor{}, nil }

// Name of the extractor
func (e Extractor) Name() string { return Name }

// Version of the extractor
func (e Extractor) Version() int { return 0 }

// Requirements of the extractor
func (e Extractor) Requirements() *plugin.Capabilities {
	return &plugin.Capabilities{}
}

// FileRequired returns true if the specified file matches uv lockfile patterns
func (e Extractor) FileRequired(api filesystem.FileAPI) bool {
	return filepath.Base(api.Path()) == "uv.lock"
}

// Extract extracts packages from uv.lock files passed through the scan input.
func (e Extractor) Extract(ctx context.Context, input *filesystem.ScanInput) (inventory.Inventory, error) {
	content, err := io.ReadAll(input.Reader)
	if err != nil {
		return inventory.Inventory{}, fmt.Errorf("could not read file: %w", err)
	}

	var parsedLockfile uvLockFile
	if err := toml.Unmarshal(content, &parsedLockfile); err != nil {
		return inventory.Inventory{}, fmt.Errorf("could not extract: %w", err)
	}

	packageNames := make([]string, 0, len(parsedLockfile.Packages))
	for _, p := range parsedLockfile.Packages {
		packageNames = append(packageNames, p.Name)
	}
	lineNums := findPackageLineNumbers(content, packageNames)

	packages := make([]*extractor.Package, 0, len(parsedLockfile.Packages))
	// Parallel to packages: the dependency names declared by packages[i].
	depNames := make([][]string, 0, len(parsedLockfile.Packages))
	var rootDeps []string

	var groups map[string][]tomldep.Dependency

	// uv stores "groups" as a table under "package" after all the packages, which due
	// to how TOML works means it ends up being a property on the last package, even
	// through in this context it's a global property rather than being per-package
	if len(parsedLockfile.Packages) > 0 {
		groups = parsedLockfile.Packages[len(parsedLockfile.Packages)-1].Groups
	}

	for i, lockPackage := range parsedLockfile.Packages {
		// skip including the root "package", since it is the subject of the scan
		// rather than one of its own dependencies
		if lockPackage.Source.isScannedProject() {
			// Its dependencies, including those behind extras and dependency
			// groups, are the project's direct dependencies.
			rootDeps = append(rootDeps, tomldep.Names(lockPackage.Dependencies)...)
			for _, group := range slices.Sorted(maps.Keys(lockPackage.Groups)) {
				rootDeps = append(rootDeps, tomldep.Names(lockPackage.Groups[group])...)
			}
			for _, group := range slices.Sorted(maps.Keys(lockPackage.DevGroups)) {
				rootDeps = append(rootDeps, tomldep.Names(lockPackage.DevGroups[group])...)
			}
			continue
		}

		_, commit, _ := strings.Cut(lockPackage.Source.Git, "#")

		pkgDetails := &extractor.Package{
			Name:     lockPackage.Name,
			Version:  lockPackage.Version,
			PURLType: purl.TypePyPi,
			Location: extractor.LocationFromPathAndLine(input.Path, lineNums[i]),
		}

		if commit != "" {
			pkgDetails.SourceCode = &extractor.SourceCodeIdentifier{
				Commit: commit,
			}
		}

		depGroupVals := []string{}

		for group, deps := range groups {
			for _, dep := range deps {
				if dep.Name == lockPackage.Name {
					depGroupVals = append(depGroupVals, group)
				}
			}
		}

		sort.Strings(depGroupVals)

		pkgDetails.Metadata = &osv.DepGroupMetadata{
			DepGroupVals: depGroupVals,
		}
		packages = append(packages, pkgDetails)
		depNames = append(depNames, tomldep.Names(lockPackage.Dependencies))
	}

	edges := depgraph.EdgesByName(packages, func(i int) []string { return depNames[i] }, pep503.Normalize)
	edges = append(edges, depgraph.RootEdgesByName(packages, rootDeps, pep503.Normalize)...)
	if err := depgraph.ApplyEdges(packages, edges); err != nil {
		return inventory.Inventory{Packages: packages}, err
	}

	return inventory.Inventory{Packages: packages}, nil
}

// extractPackageName parses a TOML key-value line and returns the unquoted
// package name if the key is "name". Returns false if the line is not a valid name assignment.
// TODO(b/491518484): Put in common location for all Python extractors to use.
func extractPackageName(line string) (string, bool) {
	if !strings.HasPrefix(line, "name") {
		return "", false
	}

	k, _, ok := strings.Cut(line, "=")
	if !ok || strings.TrimSpace(k) != "name" {
		return "", false
	}

	var pkg uvLockPackage
	if err := toml.Unmarshal([]byte(line), &pkg); err != nil {
		return "", false
	}

	return pkg.Name, true
}

// findPackageLineNumbers returns an array of line numbers that correspond to the array of package
// names passed in.
func findPackageLineNumbers(content []byte, packageNames []string) []int {
	lineNums := make([]int, len(packageNames))
	if len(packageNames) == 0 {
		return lineNums
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	pkgIdx := 0
	inPackageBlock := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "[[package]]" {
			inPackageBlock = true
			continue
		}

		if inPackageBlock && strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[[package]]") {
			inPackageBlock = false
			continue
		}

		if !inPackageBlock || pkgIdx >= len(packageNames) {
			continue
		}

		extractedName, ok := extractPackageName(line)
		if !ok || extractedName != packageNames[pkgIdx] {
			continue
		}

		lineNums[pkgIdx] = lineNum
		pkgIdx++
		inPackageBlock = false

		if pkgIdx == len(packageNames) {
			break
		}
	}

	return lineNums
}

var _ filesystem.Extractor = Extractor{}
