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

// Package pnpmlock extracts pnpm-lock.yaml files.
package pnpmlock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/internal/depgraph"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/log"
	"github.com/google/osv-scalibr/plugin"
	"github.com/google/osv-scalibr/purl"
	"gopkg.in/yaml.v3"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

const (
	// Name is the unique name of this extractor.
	Name = "javascript/pnpmlock"
)

type pnpmLockPackageResolution struct {
	Tarball string `yaml:"tarball"`
	Commit  string `yaml:"commit"`
	Repo    string `yaml:"repo"`
	Type    string `yaml:"type"`
}

type pnpmLockPackage struct {
	Resolution pnpmLockPackageResolution `yaml:"resolution"`
	Name       string                    `yaml:"name"`
	Version    string                    `yaml:"version"`
	Dev        bool                      `yaml:"dev"`
	// Dependencies of this package, in v5 and v6 lockfiles. v9 moved them to
	// the snapshots section.
	Dependencies         map[string]string `yaml:"dependencies,omitempty"`
	OptionalDependencies map[string]string `yaml:"optionalDependencies,omitempty"`
}

// pnpmSnapshot is a v9 snapshots entry, which holds the resolved dependencies
// that earlier lockfile versions kept on the package entry.
type pnpmSnapshot struct {
	Dependencies         map[string]string `yaml:"dependencies,omitempty"`
	OptionalDependencies map[string]string `yaml:"optionalDependencies,omitempty"`
}

// pnpmRootDep is a direct dependency of the project. v5 records the resolved
// version as a bare string, v6 and v9 as a {specifier, version} mapping.
type pnpmRootDep struct {
	Version string
}

// UnmarshalYAML accepts either lockfile spelling of a direct dependency.
func (d *pnpmRootDep) UnmarshalYAML(unmarshal func(any) error) error {
	var version string
	if err := unmarshal(&version); err == nil {
		d.Version = version
		return nil
	}

	var obj struct {
		Version string `yaml:"version"`
	}
	if err := unmarshal(&obj); err != nil {
		return err
	}
	d.Version = obj.Version

	return nil
}

// pnpmImporter is a workspace member's direct dependencies.
type pnpmImporter struct {
	Dependencies    map[string]pnpmRootDep `yaml:"dependencies,omitempty"`
	DevDependencies map[string]pnpmRootDep `yaml:"devDependencies,omitempty"`
}

type pnpmLockfile struct {
	Version   float64
	Packages  map[string]pnpmLockPackage
	Snapshots map[string]pnpmSnapshot
	Importers map[string]pnpmImporter
	// Root holds the top-level direct dependencies of v5 and v6 lockfiles,
	// which predate the importers section.
	Root pnpmImporter
}

type pnpmLockfileRaw struct {
	Version         string                     `yaml:"lockfileVersion"`
	Packages        map[string]pnpmLockPackage `yaml:"packages,omitempty"`
	Snapshots       map[string]pnpmSnapshot    `yaml:"snapshots,omitempty"`
	Importers       map[string]pnpmImporter    `yaml:"importers,omitempty"`
	Dependencies    map[string]pnpmRootDep     `yaml:"dependencies,omitempty"`
	DevDependencies map[string]pnpmRootDep     `yaml:"devDependencies,omitempty"`
}

// UnmarshalYAML is a custom unmarshalling function for handling the lockfile
// version, which v6 quotes as a string.
func (l *pnpmLockfile) UnmarshalYAML(unmarshal func(any) error) error {
	var raw pnpmLockfileRaw

	if err := unmarshal(&raw); err != nil {
		return err
	}

	parsedVersion, err := strconv.ParseFloat(raw.Version, 64)

	if err != nil {
		return err
	}

	l.Version = parsedVersion
	l.Packages = raw.Packages
	l.Snapshots = raw.Snapshots
	l.Importers = raw.Importers
	l.Root = pnpmImporter{Dependencies: raw.Dependencies, DevDependencies: raw.DevDependencies}

	return nil
}

var (
	numberMatcher = regexp.MustCompile(`^\d`)
	// Looks for the pattern "name@version", where name is allowed to contain zero or more "@"
	nameVersionRegexp = regexp.MustCompile(`^(.+)@([\w.-]+)(?:\(|$)`)

	codeLoadURLRegexp = regexp.MustCompile(`https://codeload\.github\.com(?:/[\w-.]+){2}/tar\.gz/(\w+)$`)
)

// extractPnpmPackageNameAndVersion parses a dependency path, attempting to
// extract the name and version of the package it represents
func extractPnpmPackageNameAndVersion(dependencyPath string, lockfileVersion float64) (string, string, error) {
	// file dependencies must always have a name property to be installed,
	// and their dependency path never has the version encoded, so we can
	// skip trying to extract either from their dependency path
	if strings.HasPrefix(dependencyPath, "file:") {
		return "", "", nil
	}

	// v9.0 specifies the dependencies as <package>@<version> rather than as a path
	if lockfileVersion >= 9.0 {
		dependencyPath = strings.Trim(dependencyPath, "'")
		dependencyPath, isScoped := strings.CutPrefix(dependencyPath, "@")

		name, version, _ := strings.Cut(dependencyPath, "@")

		if isScoped {
			name = "@" + name
		}

		return name, version, nil
	}

	parts := strings.Split(dependencyPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid dependency path: %v", dependencyPath)
	}
	var name string

	parts = parts[1:]

	if strings.HasPrefix(parts[0], "@") {
		name = strings.Join(parts[:2], "/")
		parts = parts[2:]
	} else {
		name = parts[0]
		parts = parts[1:]
	}

	version := ""

	if len(parts) != 0 {
		version = parts[0]
	}

	if version == "" {
		name, version = parseNameAtVersion(name)
	}

	if version == "" || !numberMatcher.MatchString(version) {
		return "", "", nil
	}

	underscoreIndex := strings.Index(version, "_")

	if underscoreIndex != -1 {
		version = strings.Split(version, "_")[0]
	}

	return name, version, nil
}

func parseNameAtVersion(value string) (name string, version string) {
	matches := nameVersionRegexp.FindStringSubmatch(value)

	if len(matches) != 3 {
		return name, ""
	}

	return matches[1], matches[2]
}

func parsePnpmLock(lockfile pnpmLockfile, packageLineMap map[string]int, path string) ([]*extractor.Package, error) {
	packages := make([]*extractor.Package, 0, len(lockfile.Packages))
	pkgByKey := make(map[string]*extractor.Package, len(lockfile.Packages))
	errs := []error{}

	// Iterate in key order so that extraction is deterministic.
	for _, s := range slices.Sorted(maps.Keys(lockfile.Packages)) {
		pkg := lockfile.Packages[s]
		name, version, err := extractPnpmPackageNameAndVersion(s, lockfile.Version)
		if err != nil {
			errs = append(errs, err)
			log.Errorf("failed to extract package version from %v: %v", pkg, err)
			continue
		}

		// "name" is only present if it's not in the dependency path and takes
		// priority over whatever name we think we've extracted (if any)
		if pkg.Name != "" {
			name = pkg.Name
		}

		// "version" is only present if it's not in the dependency path and takes
		// priority over whatever version we think we've extracted (if any)
		if pkg.Version != "" {
			version = pkg.Version
		}

		if name == "" || version == "" {
			continue
		}

		commit := pkg.Resolution.Commit

		if strings.HasPrefix(pkg.Resolution.Tarball, "https://codeload.github.com") {
			matched := codeLoadURLRegexp.FindStringSubmatch(pkg.Resolution.Tarball)

			if matched != nil {
				commit = matched[1]
			}
		}

		depGroups := []string{}
		if pkg.Dev {
			depGroups = append(depGroups, "dev")
		}

		lineNum := packageLineMap[s]
		p := &extractor.Package{
			Name:     name,
			Version:  version,
			PURLType: purl.TypeNPM,
			SourceCode: &extractor.SourceCodeIdentifier{
				Commit: commit,
			},
			Metadata: &osv.DepGroupMetadata{
				DepGroupVals: depGroups,
			},
			Location: extractor.LocationFromPathAndLine(path, lineNum),
		}
		packages = append(packages, p)
		pkgByKey[s] = p
	}

	if err := depgraph.ApplyEdges(packages, pnpmEdges(lockfile, pkgByKey)); err != nil {
		errs = append(errs, err)
	}

	return packages, errors.Join(errs...)
}

// pnpmEdges resolves the lockfile's dependency declarations against the keys of
// the emitted packages. pnpm dependency paths are globally unique, so a
// declaration names its instance outright and no install-tree walk is needed.
func pnpmEdges(lockfile pnpmLockfile, pkgByKey map[string]*extractor.Package) []depgraph.Edge {
	var edges []depgraph.Edge

	addEdges := func(parent *extractor.Package, deps ...map[string]string) {
		for _, m := range deps {
			for _, name := range slices.Sorted(maps.Keys(m)) {
				if child, ok := pkgByKey[pnpmDepKey(name, m[name], lockfile.Version)]; ok && child != parent {
					edges = append(edges, depgraph.Edge{Parent: parent, Child: child})
				}
			}
		}
	}

	// v9 records resolved dependencies under snapshots, keyed like the packages
	// entries but with the peer suffix retained.
	for _, key := range slices.Sorted(maps.Keys(lockfile.Snapshots)) {
		snapshot := lockfile.Snapshots[key]
		addEdges(pkgByKey[stripPnpmPeerSuffix(key)], snapshot.Dependencies, snapshot.OptionalDependencies)
	}

	// v5 and v6 record them on the package entry itself.
	for _, key := range slices.Sorted(maps.Keys(lockfile.Packages)) {
		pkg := lockfile.Packages[key]
		addEdges(pkgByKey[key], pkg.Dependencies, pkg.OptionalDependencies)
	}

	// Direct dependencies: importers in v6 and v9, the top level in v5. Only
	// the root importer is a direct dependency of the scanned project.
	roots := []pnpmImporter{lockfile.Root}
	if root, ok := lockfile.Importers["."]; ok {
		roots = append(roots, root)
	}
	for _, importer := range roots {
		for _, m := range []map[string]pnpmRootDep{importer.Dependencies, importer.DevDependencies} {
			for _, name := range slices.Sorted(maps.Keys(m)) {
				if child, ok := pkgByKey[pnpmDepKey(name, m[name].Version, lockfile.Version)]; ok {
					edges = append(edges, depgraph.Edge{Child: child})
				}
			}
		}
	}

	return edges
}

// pnpmDepKey builds the packages-section key that a name and resolved version
// refer to. v9 keys read "name@version", earlier ones "/name/version" (v5) and
// "/name@version" (v6).
func pnpmDepKey(name, version string, lockfileVersion float64) string {
	switch {
	case lockfileVersion >= 9.0:
		return name + "@" + stripPnpmPeerSuffix(version)
	case lockfileVersion >= 6.0:
		return "/" + name + "@" + version
	default:
		return "/" + name + "/" + version
	}
}

// stripPnpmPeerSuffix removes the "(peer@version)" suffixes that v9 appends to
// snapshot keys and resolved versions; the packages section omits them.
func stripPnpmPeerSuffix(s string) string {
	if i := strings.Index(s, "("); i != -1 {
		return s[:i]
	}
	return s
}

// Extractor extracts pnpm-lock.yaml files.
type Extractor struct{}

// New returns a new instance of the extractor.
func New(_ *cpb.PluginConfig) (filesystem.Extractor, error) { return &Extractor{}, nil }

// Name of the extractor
func (e Extractor) Name() string { return Name }

// Version of the extractor
func (e Extractor) Version() int { return 0 }

// Requirements of the extractor.
func (e Extractor) Requirements() *plugin.Capabilities { return &plugin.Capabilities{} }

// FileRequired returns true if the specified file matches pnpm-lock.yaml files.
func (e Extractor) FileRequired(api filesystem.FileAPI) bool {
	path := api.Path()
	if filepath.Base(path) != "pnpm-lock.yaml" {
		return false
	}
	// Skip lockfiles inside node_modules directories since the packages they list aren't
	// necessarily installed by the root project. We instead use the more specific top-level
	// lockfile for the root project dependencies.
	dir := filepath.ToSlash(filepath.Dir(path))
	return !slices.Contains(strings.Split(dir, "/"), "node_modules")
}

// Extract extracts packages from a pnpm-lock.yaml file passed through the scan input.
func (e Extractor) Extract(ctx context.Context, input *filesystem.ScanInput) (inventory.Inventory, error) {
	var root yaml.Node
	if err := yaml.NewDecoder(input.Reader).Decode(&root); err != nil {
		if errors.Is(err, io.EOF) {
			return inventory.Inventory{Packages: []*extractor.Package{}}, nil
		}
		return inventory.Inventory{}, fmt.Errorf("could not extract: %w", err)
	}

	var parsedLockfile pnpmLockfile
	if err := root.Decode(&parsedLockfile); err != nil {
		return inventory.Inventory{}, fmt.Errorf("could not extract: %w", err)
	}

	packageLineMap := findLineNumbers(&root)

	packages, err := parsePnpmLock(parsedLockfile, packageLineMap, input.Path)
	return inventory.Inventory{Packages: packages}, err
}

// findLineNumbers goes through the Node tree to find the line numbers for each package.
func findLineNumbers(root *yaml.Node) map[string]int {
	results := make(map[string]int)
	if len(root.Content) == 0 {
		return results
	}
	doc := root.Content[0]

	if doc.Kind != yaml.MappingNode {
		return results // empty results
	}

	var packagesNode *yaml.Node
	// Note: increment by 2 to iterate from key to key (skip the value).
	for i := 0; i < len(doc.Content); i += 2 {
		if doc.Content[i].Value == "packages" {
			packagesNode = doc.Content[i+1]
			break
		}
	}

	if packagesNode == nil || packagesNode.Kind != yaml.MappingNode {
		return results // empty results
	}
	// Note: increment by 2 to iterate from key to key (skip the value).
	for i := 0; i < len(packagesNode.Content); i += 2 {
		keyNode := packagesNode.Content[i]
		results[keyNode.Value] = keyNode.Line
	}
	return results
}

var _ filesystem.Extractor = Extractor{}
