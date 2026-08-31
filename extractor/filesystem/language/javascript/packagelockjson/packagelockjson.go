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

// Package packagelockjson extracts package-lock.json files.
package packagelockjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/internal/depgraph"
	"github.com/google/osv-scalibr/extractor/filesystem/internal/linefinder"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/internal/commitextractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/internal/depkey"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/internal/rootdeps"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/internal/dependencyfile/packagelockjson"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/plugin"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/stats"
	"github.com/tidwall/gjson"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

const (
	// Name is the unique name of this extractor.
	Name = "javascript/packagelockjson"

	// noLimitMaxFileSizeBytes is a sentinel value that indicates no limit.
	noLimitMaxFileSizeBytes = int64(0)
)

type packageDetails struct {
	Name      string
	Version   string
	Commit    string
	DepGroups []string
	Line      int
}

type npmPackageDetailsMap map[string]packageDetails

// lockGraph records the install tree alongside the deduplicated package list.
// Packages are deduplicated by "name@version" because two install locations of
// one version are the same package, but resolving a dependency needs the
// install path it was declared at, so both are kept.
type lockGraph struct {
	// keyByPath maps an install path ("node_modules/a/node_modules/b") to the
	// deduplicated package key ("b@1.0.0").
	keyByPath map[string]string
	// depsByPath lists the dependency names declared at an install path.
	depsByPath map[string][]string
	// rootDeps are the names the project depends on directly, when the
	// lockfile records them.
	rootDeps []string
	// linkTargets maps a symlink install path ("node_modules/w") to the path
	// of the package it points at ("packages/w").
	linkTargets map[string]string
}

func newLockGraph() *lockGraph {
	return &lockGraph{keyByPath: map[string]string{}, depsByPath: map[string][]string{}, linkTargets: map[string]string{}}
}

// resolveLinks points each symlink install path at the key of the package it
// links to, so that a dependency resolving through the symlink lands on the
// real package. Links whose target is not itself a package entry are dropped.
func (g *lockGraph) resolveLinks() {
	for linkPath, target := range g.linkTargets {
		if key, ok := g.keyByPath[target]; ok {
			g.keyByPath[linkPath] = key
		}
	}
}

func (g *lockGraph) add(installPath, key string, deps ...map[string]string) {
	g.keyByPath[installPath] = key
	g.depsByPath[installPath] = mergeDepNames(deps...)
}

// mergeDepNames returns the sorted union of the keys of deps.
func mergeDepNames(deps ...map[string]string) []string {
	seen := map[string]bool{}
	for _, m := range deps {
		for name := range m {
			seen[name] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

// mergeNpmDepsGroups handles merging the dependency groups of packages within the
// NPM ecosystem, since they can appear multiple times in the same dependency tree
//
// the merge happens almost as you'd expect, except that if either given packages
// belong to no groups, then that is the result since it indicates the package
// is implicitly a production dependency.
func mergeNpmDepsGroups(a, b packageDetails) []string {
	// if either group includes no groups, then the package is in the "production" group
	if len(a.DepGroups) == 0 || len(b.DepGroups) == 0 {
		return nil
	}

	combined := make([]string, 0, len(a.DepGroups)+len(b.DepGroups))
	combined = append(combined, a.DepGroups...)
	combined = append(combined, b.DepGroups...)

	slices.Sort(combined)

	return slices.Compact(combined)
}

func (pdm npmPackageDetailsMap) add(key string, details packageDetails) {
	existing, ok := pdm[key]

	if ok {
		details.DepGroups = mergeNpmDepsGroups(existing, details)
	}

	pdm[key] = details
}

func parseNpmLockDependencies(dependencies map[string]packagelockjson.Dependency, finder *linefinder.JSONLineFinder, parentPath, installPath string, graph *lockGraph) map[string]packageDetails {
	details := npmPackageDetailsMap{}

	for name, detail := range dependencies {
		currentPath := parentPath + "." + gjson.Escape(name)
		// The install directory is keyed by the entry name even when the
		// package is aliased, so capture it before any alias rewrite below.
		currentInstall := installPath + "node_modules/" + name

		if detail.Dependencies != nil {
			nestedDeps := parseNpmLockDependencies(detail.Dependencies, finder, currentPath+".dependencies", currentInstall+"/", graph)
			for k, v := range nestedDeps {
				details.add(k, v)
			}
		}

		version := detail.Version
		finalVersion := version
		commit := ""

		// If the package is aliased, get the name and version
		// E.g. npm:string-width@^4.2.0
		if strings.HasPrefix(detail.Version, "npm:") {
			i := strings.LastIndex(detail.Version, "@")
			name = detail.Version[4:i]
			finalVersion = detail.Version[i+1:]
		}

		// we can't resolve a version from a "file:" dependency
		if strings.HasPrefix(detail.Version, "file:") {
			finalVersion = ""
		} else {
			commit = commitextractor.TryExtractCommit(detail.Version)

			// if there is a commit, we want to deduplicate based on that rather than
			// the version (the versions must match anyway for the commits to match)
			//
			// we also don't actually know what the "version" is, so blank it
			if commit != "" {
				finalVersion = ""
				version = commit
			}
		}

		line := 0
		if finder != nil {
			line = finder.LineOf(currentPath)
		}

		details.add(name+"@"+version, packageDetails{
			Name:      name,
			Version:   finalVersion,
			Commit:    commit,
			DepGroups: detail.DepGroups(),
			Line:      line,
		})
		graph.add(currentInstall, name+"@"+version, detail.Requires)
	}

	return details
}

func extractNpmPackageName(name string) string {
	maybeScope := path.Base(path.Dir(name))
	pkgName := path.Base(name)

	if strings.HasPrefix(maybeScope, "@") {
		pkgName = maybeScope + "/" + pkgName
	}

	return pkgName
}

func parseNpmLockPackages(packages map[string]packagelockjson.Package, finder *linefinder.JSONLineFinder, graph *lockGraph) map[string]packageDetails {
	details := npmPackageDetailsMap{}

	for namePath, detail := range packages {
		if namePath == "" {
			// The project itself. It is not one of its own dependencies, but
			// it names them.
			graph.rootDeps = mergeDepNames(detail.Dependencies, detail.DevDependencies, detail.OptionalDependencies, detail.PeerDependencies)
			continue
		}

		if detail.Link {
			// A symlink to a workspace member, which has its own entry under
			// the path this resolves to.
			graph.linkTargets[namePath] = detail.Resolved
			continue
		}

		finalName := detail.Name
		if finalName == "" {
			finalName = extractNpmPackageName(namePath)
		}

		finalVersion := detail.Version

		commit := commitextractor.TryExtractCommit(detail.Resolved)

		// if there is a commit, we want to deduplicate based on that rather than
		// the version (the versions must match anyway for the commits to match)
		if commit != "" {
			finalVersion = commit
		}

		line := 0
		if finder != nil {
			line = finder.LineOf("packages." + gjson.Escape(namePath))
		}

		details.add(finalName+"@"+finalVersion, packageDetails{
			Name:      finalName,
			Version:   detail.Version,
			Commit:    commit,
			DepGroups: detail.DepGroups(),
			Line:      line,
		})
		graph.add(namePath, finalName+"@"+finalVersion,
			detail.Dependencies, detail.DevDependencies, detail.OptionalDependencies, detail.PeerDependencies)
	}

	graph.resolveLinks()

	return details
}

func parseNpmLock(lockfile packagelockjson.LockFile, finder *linefinder.JSONLineFinder, graph *lockGraph) map[string]packageDetails {
	if lockfile.Packages != nil {
		return parseNpmLockPackages(lockfile.Packages, finder, graph)
	}

	return parseNpmLockDependencies(lockfile.Dependencies, finder, "dependencies", "", graph)
}

// Extractor extracts npm packages from package-lock.json files.
type Extractor struct {
	Stats            stats.Collector
	maxFileSizeBytes int64
}

// New returns a package-lock.json extractor.
func New(cfg *cpb.PluginConfig) (filesystem.Extractor, error) {
	maxFileSizeBytes := noLimitMaxFileSizeBytes
	if cfg.GetMaxFileSizeBytes() > 0 {
		maxFileSizeBytes = cfg.GetMaxFileSizeBytes()
	}

	specific := plugin.FindConfig(cfg, func(c *cpb.PluginSpecificConfig) *cpb.JavascriptPackageLockJsonConfig {
		return c.GetJavascriptPackageLockJson()
	})
	if specific.GetMaxFileSizeBytes() > 0 {
		maxFileSizeBytes = specific.GetMaxFileSizeBytes()
	}

	return &Extractor{maxFileSizeBytes: maxFileSizeBytes}, nil
}

// Name of the extractor.
func (e Extractor) Name() string { return Name }

// Version of the extractor.
func (e Extractor) Version() int { return 0 }

// Requirements of the extractor.
func (e Extractor) Requirements() *plugin.Capabilities {
	return &plugin.Capabilities{}
}

// FileRequired returns true if the specified file matches npm lockfile patterns.
func (e Extractor) FileRequired(api filesystem.FileAPI) bool {
	path := api.Path()
	if !slices.Contains([]string{"package-lock.json", "npm-shrinkwrap.json"}, filepath.Base(path)) {
		return false
	}
	// Skip lockfiles inside node_modules directories since the packages they list aren't
	// necessarily installed by the root project. We instead use the more specific top-level
	// lockfile for the root project dependencies.
	dir := filepath.ToSlash(filepath.Dir(path))
	if slices.Contains(strings.Split(dir, "/"), "node_modules") {
		return false
	}

	fileInfo, err := api.Stat()
	if err != nil {
		return false
	}
	if e.maxFileSizeBytes > noLimitMaxFileSizeBytes && fileInfo.Size() > e.maxFileSizeBytes {
		e.reportFileRequired(path, fileInfo.Size(), stats.FileRequiredResultSizeLimitExceeded)
		return false
	}

	e.reportFileRequired(path, fileInfo.Size(), stats.FileRequiredResultOK)
	return true
}

func (e Extractor) reportFileRequired(path string, fileSizeBytes int64, result stats.FileRequiredResult) {
	if e.Stats == nil {
		return
	}
	e.Stats.AfterFileRequired(e.Name(), &stats.FileRequiredStats{
		Path:          path,
		Result:        result,
		FileSizeBytes: fileSizeBytes,
	})
}

// Extract extracts packages from package-lock.json files passed through the scan input.
func (e Extractor) Extract(ctx context.Context, input *filesystem.ScanInput) (inventory.Inventory, error) {
	packages, err := e.extractPkgLock(ctx, input)

	if e.Stats != nil {
		var fileSizeBytes int64
		if input.Info != nil {
			fileSizeBytes = input.Info.Size()
		}
		e.Stats.AfterFileExtracted(e.Name(), &stats.FileExtractedStats{
			Path:          input.Path,
			Result:        filesystem.ExtractorErrorToFileExtractedResult(err),
			FileSizeBytes: fileSizeBytes,
		})
	}

	return inventory.Inventory{Packages: packages}, err
}

func (e Extractor) extractPkgLock(_ context.Context, input *filesystem.ScanInput) ([]*extractor.Package, error) {
	// If both package-lock.json and npm-shrinkwrap.json are present in the root of a project,
	// npm-shrinkwrap.json will take precedence and package-lock.json will be ignored.
	if filepath.Base(input.Path) == "package-lock.json" {
		npmShrinkwrapPath := path.Join(filepath.ToSlash(filepath.Dir(input.Path)), "npm-shrinkwrap.json")
		_, err := input.FS.Open(npmShrinkwrapPath)
		if err == nil {
			return nil, nil
		}
	}

	b, err := io.ReadAll(input.Reader)
	if err != nil {
		return nil, fmt.Errorf("could not read: %w", err)
	}

	var parsedLockfile *packagelockjson.LockFile
	if err := json.Unmarshal(b, &parsedLockfile); err != nil {
		return nil, fmt.Errorf("could not extract: %w", err)
	}

	if parsedLockfile == nil {
		return nil, errors.New("could not extract: decoded null JSON value")
	}

	finder := linefinder.NewJSONLineFinder(b)

	graph := newLockGraph()
	details := parseNpmLock(*parsedLockfile, finder, graph)

	// Iterate in key order so that extraction is deterministic.
	keys := slices.Sorted(maps.Keys(details))
	result := make([]*extractor.Package, len(keys))
	pkgByKey := make(map[string]*extractor.Package, len(keys))

	for i, key := range keys {
		pkg := details[key]
		if pkg.DepGroups == nil {
			pkg.DepGroups = []string{}
		}

		result[i] = &extractor.Package{
			Name: pkg.Name,
			SourceCode: &extractor.SourceCodeIdentifier{
				Commit: pkg.Commit,
			},
			Version:  pkg.Version,
			PURLType: purl.TypeNPM,
			Metadata: &osv.DepGroupMetadata{
				DepGroupVals: pkg.DepGroups,
			},
			Location: extractor.LocationFromPathAndLine(input.Path, pkg.Line),
		}
		pkgByKey[key] = result[i]
	}

	// v1 lockfiles do not record which entries are direct dependencies; the
	// project's own manifest is the only source for that.
	if graph.rootDeps == nil {
		graph.rootDeps = slices.Sorted(maps.Keys(rootdeps.FromSiblingPackageJSON(input)))
	}

	if err := depgraph.ApplyEdges(result, npmEdges(graph, pkgByKey)); err != nil {
		return result, err
	}

	return result, nil
}

// npmEdges resolves each declared dependency to the install it refers to,
// following npm's nested resolution, then maps that install back to the
// deduplicated package it belongs to.
func npmEdges(graph *lockGraph, pkgByKey map[string]*extractor.Package) []depgraph.Edge {
	paths := make(map[string]bool, len(graph.keyByPath))
	for p := range graph.keyByPath {
		paths[p] = true
	}
	resolver := depkey.Resolver{Segment: "node_modules/", Keys: paths}

	child := func(fromPath, depName string) (*extractor.Package, bool) {
		childPath, ok := resolver.Resolve(fromPath, depName)
		if !ok {
			return nil, false
		}
		pkg, ok := pkgByKey[graph.keyByPath[childPath]]
		return pkg, ok
	}

	var edges []depgraph.Edge
	for _, installPath := range slices.Sorted(maps.Keys(graph.depsByPath)) {
		parent := pkgByKey[graph.keyByPath[installPath]]
		for _, depName := range graph.depsByPath[installPath] {
			if c, ok := child(installPath, depName); ok && c != parent {
				edges = append(edges, depgraph.Edge{Parent: parent, Child: c})
			}
		}
	}
	for _, depName := range graph.rootDeps {
		if c, ok := child("", depName); ok {
			edges = append(edges, depgraph.Edge{Child: c})
		}
	}

	return edges
}
