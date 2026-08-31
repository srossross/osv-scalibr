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

package depkey_test

import (
	"testing"

	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/internal/depkey"
)

func TestResolve_bunSegment(t *testing.T) {
	r := depkey.Resolver{Keys: map[string]bool{
		"a":                    true,
		"a/b":                  true,
		"a/b/c":                true,
		"c":                    true,
		"has-flag":             true,
		"@types/react-dom":     true,
		"@types/prop-types":    true,
		"x/@babel/core":        true,
		"x/@babel/core/nested": true,
	}}
	tests := []struct {
		name            string
		pkgKey, depName string
		want            string
		wantOK          bool
	}{
		{name: "nested_match", pkgKey: "a/b", depName: "c", want: "a/b/c", wantOK: true},
		{name: "walk_up_to_top_level", pkgKey: "a/b", depName: "has-flag", want: "has-flag", wantOK: true},
		{name: "top_level", pkgKey: "a", depName: "c", want: "c", wantOK: true},
		{name: "miss", pkgKey: "a/b", depName: "missing", want: "", wantOK: false},
		{name: "self_edge_blocked", pkgKey: "c", depName: "c", want: "", wantOK: false},
		{name: "scoped_dep_from_root", pkgKey: "", depName: "@types/prop-types", want: "@types/prop-types", wantOK: true},
		{
			// Bare "prop-types" from a scoped parent must not match
			// "@types/prop-types" via a half-stripped "@types" prefix.
			name: "scoped_parent_does_not_leak_scope_prefix", pkgKey: "@types/react-dom", depName: "prop-types",
			want: "", wantOK: false,
		},
		{name: "nested_under_scoped_segment", pkgKey: "x/@babel/core", depName: "nested", want: "x/@babel/core/nested", wantOK: true},
		{name: "scoped_segment_stripped_whole", pkgKey: "x/@babel/core", depName: "has-flag", want: "has-flag", wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := r.Resolve(tt.pkgKey, tt.depName)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Resolve(%q, %q) = (%q, %v), want (%q, %v)", tt.pkgKey, tt.depName, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestResolve_nodeModulesSegment(t *testing.T) {
	r := depkey.Resolver{Segment: "node_modules/", Keys: map[string]bool{
		"node_modules/a":                               true,
		"node_modules/a/node_modules/b":                true,
		"node_modules/a/node_modules/b/node_modules/c": true,
		"node_modules/c":                               true,
		"node_modules/has-flag":                        true,
		"node_modules/@types/react-dom":                true,
		"node_modules/a/node_modules/@babel/core":      true,
	}}
	tests := []struct {
		name            string
		pkgKey, depName string
		want            string
		wantOK          bool
	}{
		{name: "nested_match", pkgKey: "node_modules/a/node_modules/b", depName: "c", want: "node_modules/a/node_modules/b/node_modules/c", wantOK: true},
		{name: "walk_up_to_top_level", pkgKey: "node_modules/a/node_modules/b", depName: "has-flag", want: "node_modules/has-flag", wantOK: true},
		{name: "from_root", pkgKey: "", depName: "a", want: "node_modules/a", wantOK: true},
		{name: "root_scoped", pkgKey: "", depName: "@types/react-dom", want: "node_modules/@types/react-dom", wantOK: true},
		{name: "miss", pkgKey: "node_modules/a", depName: "missing", want: "", wantOK: false},
		{name: "self_edge_blocked", pkgKey: "node_modules/c", depName: "c", want: "", wantOK: false},
		{name: "nested_scoped", pkgKey: "node_modules/a", depName: "@babel/core", want: "node_modules/a/node_modules/@babel/core", wantOK: true},
		{name: "scoped_key_walks_to_top", pkgKey: "node_modules/@types/react-dom", depName: "has-flag", want: "node_modules/has-flag", wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := r.Resolve(tt.pkgKey, tt.depName)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Resolve(%q, %q) = (%q, %v), want (%q, %v)", tt.pkgKey, tt.depName, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
