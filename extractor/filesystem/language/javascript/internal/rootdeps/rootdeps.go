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

// Package rootdeps reads the direct dependencies a project declares, for
// lockfile formats that do not record which of their entries are direct.
package rootdeps

import (
	"encoding/json"
	"io"
	"path"
	"path/filepath"

	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/log"
)

type packageJSON struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

// FromSiblingPackageJSON returns the name->range pairs declared by the
// package.json next to the scanned lockfile. A missing or unreadable
// package.json yields no dependencies rather than an error: the lockfile is
// still worth extracting, it just has no direct/transitive distinction.
//
// In a workspace this is the workspace root's manifest, so it names only the
// root's own dependencies. Formats that record direct dependencies themselves
// (package-lock.json v2+, pnpm-lock.yaml) should use those instead.
func FromSiblingPackageJSON(input *filesystem.ScanInput) map[string]string {
	if input.FS == nil {
		return nil
	}

	p := path.Join(filepath.ToSlash(filepath.Dir(input.Path)), "package.json")
	f, err := input.FS.Open(p)
	if err != nil {
		log.Debugf("no sibling package.json for %s: %v", input.Path, err)
		return nil
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Debugf("could not read %s: %v", p, err)
		return nil
	}

	var manifest packageJSON
	if err := json.Unmarshal(b, &manifest); err != nil {
		log.Debugf("could not parse %s: %v", p, err)
		return nil
	}

	deps := make(map[string]string, len(manifest.Dependencies)+len(manifest.DevDependencies)+len(manifest.OptionalDependencies))
	for _, m := range []map[string]string{manifest.Dependencies, manifest.DevDependencies, manifest.OptionalDependencies} {
		for name, constraint := range m {
			deps[name] = constraint
		}
	}
	return deps
}
