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

// Package yarnlock extracts NPC yarn.lock files.
package yarnlock

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/internal/depgraph"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/internal/commitextractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/internal/rootdeps"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/log"
	"github.com/google/osv-scalibr/plugin"
	"github.com/google/osv-scalibr/purl"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

const (
	// Name is the unique name of this extractor.
	Name = "javascript/yarnlock"
)

var (
	// Version matcher regex.
	// Format for yarn.lock v1: `version "0.0.1"`
	// Format for yarn.lock v2: `version: 0.0.1`
	yarnPackageVersionRe = regexp.MustCompile(`^ {2}"?version"?:? "?([\w-.+]+)"?$`)
	// Package resolution matcher regex. Might contain commit hashes.
	// Format for yarn.lock v1: `resolved "git+ssh://git@github.com:G-Rath/repo-2#hash"`
	// Format for yarn.lock v2: `resolution: "@my-scope/my-first-package@https://github.com/my-org/my-first-pkg.git#commit=hash"`
	yarnPackageResolutionRe = regexp.MustCompile(`^ {2}"?(?:resolution:|resolved)"? "([^ '"]+)"$`)
	// Start of a dependencies block within an entry, in either lockfile format.
	yarnDepBlockRe = regexp.MustCompile(`^"?(dependencies|optionalDependencies|peerDependencies)"?:$`)
)

func shouldSkipYarnLine(line string) bool {
	line = strings.TrimSpace(line)
	return line == "" || strings.HasPrefix(line, "#")
}

// yaml.lock files define packages as follows:
//
//	header
//	  prop1 value1
//	  prop2 value2
//
//	header2
//	  prop3 value3
type packageDescription struct {
	header     string
	props      []string
	lineNumber int
}

func groupYarnPackageDescriptions(ctx context.Context, scanner *bufio.Scanner) ([]*packageDescription, error) {
	var result []*packageDescription

	var current *packageDescription
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if err := scanner.Err(); err != nil {
			return result, err
		}

		line := scanner.Text()

		if shouldSkipYarnLine(line) {
			continue
		}

		// represents the start of a new dependency
		if !strings.HasPrefix(line, " ") {
			// Add previous descriptor if it's for a package.
			if current != nil {
				result = append(result, current)
			}
			current = &packageDescription{header: line, lineNumber: lineNumber}
		} else if current == nil {
			return nil, errors.New("malformed yarn.lock")
		} else {
			current.props = append(current.props, line)
		}
	}
	// Add trailing descriptor.
	if current != nil {
		result = append(result, current)
	}

	return result, nil
}

func extractYarnPackageName(header string) string {
	// Header format: @my-scope/my-first-package@my-scope/my-first-package#commit=hash
	str := strings.TrimPrefix(header, "\"")
	str = strings.TrimSuffix(str, ":")
	str, _, _ = strings.Cut(str, ",")

	isScoped := strings.HasPrefix(str, "@")

	if isScoped {
		str = strings.TrimPrefix(str, "@")
	}
	name, right, _ := strings.Cut(str, "@")

	// Packages can also contain an NPM entry, e.g. @nicolo-ribaudo/chokidar-2@npm:2.1.8-no-fsevents.3
	if strings.HasPrefix(right, "npm:") && strings.Contains(right, "@") {
		return extractYarnPackageName(strings.TrimPrefix(right, "npm:"))
	}

	if isScoped {
		name = "@" + name
	}
	return name
}

func determineYarnPackageVersion(props []string) string {
	for _, s := range props {
		matched := yarnPackageVersionRe.FindStringSubmatch(s)

		if matched != nil {
			return matched[1]
		}
	}
	return ""
}

func determineYarnPackageResolution(props []string) string {
	for _, s := range props {
		matched := yarnPackageResolutionRe.FindStringSubmatch(s)
		if matched != nil {
			return matched[1]
		}
	}
	return ""
}

// yarnDescriptors returns the descriptors an entry header declares, e.g.
// `"a@npm:^1.0.0", "a@npm:^1.2.0":` yields both. Each is registered under its
// literal form and, for berry, under the protocol-less form the dependency
// blocks use.
func yarnDescriptors(header string) []string {
	header = strings.TrimSuffix(strings.TrimSpace(header), ":")
	var out []string
	for _, d := range strings.Split(header, ",") {
		d = strings.Trim(strings.TrimSpace(d), `"`)
		if d == "" {
			continue
		}
		out = append(out, d)
		if name, constraint, ok := strings.Cut(d, "@npm:"); ok {
			out = append(out, name+"@"+constraint)
		}
	}
	return out
}

// yarnDependencies returns the name->constraint pairs of an entry's
// dependencies blocks. v1 writes `    name "range"`, berry `    name: range`.
func yarnDependencies(props []string) map[string]string {
	deps := map[string]string{}
	inBlock := false
	for _, line := range props {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		trimmed := strings.TrimSpace(line)
		if indent <= 2 {
			inBlock = yarnDepBlockRe.MatchString(trimmed)
			continue
		}
		if !inBlock {
			continue
		}
		name, constraint, ok := strings.Cut(trimmed, ": ")
		if !ok {
			name, constraint, ok = strings.Cut(trimmed, " ")
			if !ok {
				continue
			}
		}
		name = strings.Trim(strings.TrimSpace(name), `"`)
		constraint = strings.Trim(strings.TrimSpace(constraint), `"`)
		if name != "" && constraint != "" {
			deps[name] = constraint
		}
	}
	return deps
}

func parseYarnPackageGroup(desc *packageDescription) *extractor.Package {
	name := extractYarnPackageName(desc.header)
	version := determineYarnPackageVersion(desc.props)
	resolution := determineYarnPackageResolution(desc.props)

	if version == "" {
		log.Errorf("Failed to determine version of %s while parsing a yarn.lock", name)
	}

	return &extractor.Package{
		Name:     name,
		Version:  version,
		PURLType: purl.TypeNPM,
		SourceCode: &extractor.SourceCodeIdentifier{
			Commit: commitextractor.TryExtractCommit(resolution),
		},
	}
}

// Extractor extracts NPM yarn.lock files.
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

// FileRequired returns true if the specified file is an NPM yarn.lock file.
func (e Extractor) FileRequired(api filesystem.FileAPI) bool {
	path := api.Path()
	if filepath.Base(path) != "yarn.lock" {
		return false
	}
	// Skip lockfiles inside node_modules directories since the packages they list aren't
	// necessarily installed by the root project. We instead use the more specific top-level
	// lockfile for the root project dependencies.
	dir := filepath.ToSlash(filepath.Dir(path))
	return !slices.Contains(strings.Split(dir, "/"), "node_modules")
}

// Extract extracts packages from NPM yarn.lock files passed through the scan input.
func (e Extractor) Extract(ctx context.Context, input *filesystem.ScanInput) (inventory.Inventory, error) {
	scanner := bufio.NewScanner(input.Reader)

	packageGroups, err := groupYarnPackageDescriptions(ctx, scanner)
	if err != nil {
		return inventory.Inventory{}, fmt.Errorf("error while scanning: %w", err)
	}

	packages := make([]*extractor.Package, 0, len(packageGroups))
	// Parallel to packages: the dependencies declared by packages[i].
	deps := make([]map[string]string, 0, len(packageGroups))
	// Every descriptor an entry answers to, mapped to that entry's package.
	byDescriptor := map[string]*extractor.Package{}

	for _, group := range packageGroups {
		if group.header == "__metadata:" {
			// This group doesn't describe a package.
			continue
		}
		if strings.HasSuffix(group.header, "@workspace:.\":") {
			// This is the root package itself.
			continue
		}
		pkg := parseYarnPackageGroup(group)
		pkg.Location = extractor.LocationFromPathAndLine(input.Path, group.lineNumber)
		packages = append(packages, pkg)
		deps = append(deps, yarnDependencies(group.props))
		for _, d := range yarnDescriptors(group.header) {
			byDescriptor[d] = pkg
		}
	}

	edges := descriptorEdges(packages, deps, byDescriptor)
	// yarn.lock does not record which entries are direct dependencies, so the
	// project's own manifest is the only source for root edges.
	for name, constraint := range rootdeps.FromSiblingPackageJSON(input) {
		if child, ok := resolveDescriptor(byDescriptor, name, constraint); ok {
			edges = append(edges, depgraph.Edge{Child: child})
		}
	}

	if err := depgraph.ApplyEdges(packages, edges); err != nil {
		return inventory.Inventory{Packages: packages}, err
	}

	return inventory.Inventory{Packages: packages}, nil
}

// descriptorEdges resolves each entry's declared dependencies to the entry that
// answers to that descriptor. Descriptors that match no entry are dropped.
func descriptorEdges(packages []*extractor.Package, deps []map[string]string, byDescriptor map[string]*extractor.Package) []depgraph.Edge {
	var edges []depgraph.Edge
	for i, parent := range packages {
		for _, name := range slices.Sorted(maps.Keys(deps[i])) {
			if child, ok := resolveDescriptor(byDescriptor, name, deps[i][name]); ok && child != parent {
				edges = append(edges, depgraph.Edge{Parent: parent, Child: child})
			}
		}
	}
	return edges
}

// resolveDescriptor looks up name at constraint, trying the berry protocol form
// for entries whose descriptors carry one.
func resolveDescriptor(byDescriptor map[string]*extractor.Package, name, constraint string) (*extractor.Package, bool) {
	if pkg, ok := byDescriptor[name+"@"+constraint]; ok {
		return pkg, true
	}
	pkg, ok := byDescriptor[name+"@npm:"+constraint]
	return pkg, ok
}

var _ filesystem.Extractor = Extractor{}
