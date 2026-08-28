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

// Package tomldep models the dependency entries that Python TOML lockfiles
// declare as inline tables, e.g. `dependencies = [{ name = "requests" }]` in
// uv.lock and pylock.toml.
package tomldep

// Dependency is a single dependency entry. Only the name is used for graph
// edges; the remaining keys (version, source, marker) vary by lockfile.
type Dependency struct {
	Name string `toml:"name"`
}

// Names returns the names of deps, in order.
func Names(deps []Dependency) []string {
	names := make([]string, 0, len(deps))
	for _, d := range deps {
		names = append(names, d.Name)
	}
	return names
}
