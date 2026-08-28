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

package bunlock

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPackageDependencies(t *testing.T) {
	tests := []struct {
		name                             string
		pkgs                             []any
		wantDeps, wantOptional, wantPeer map[string]string
	}{
		{
			name: "registry_tuple",
			pkgs: []any{"a@1.0.0", "", map[string]any{
				"dependencies":         map[string]any{"b": "^1.0.0"},
				"optionalDependencies": map[string]any{"c": "^2.0.0"},
				"peerDependencies":     map[string]any{"d": "*"},
			}, "sha512-..."},
			wantDeps:     map[string]string{"b": "^1.0.0"},
			wantOptional: map[string]string{"c": "^2.0.0"},
			wantPeer:     map[string]string{"d": "*"},
		},
		{
			name:     "git_tuple_with_metadata_at_index_1",
			pkgs:     []any{"a@github:o/r#abc", map[string]any{"dependencies": map[string]any{"b": "^1.0.0"}}, "marker"},
			wantDeps: map[string]string{"b": "^1.0.0"},
		},
		{
			name: "file_tuple_with_empty_metadata",
			pkgs: []any{"a@file:deps/a", map[string]any{}},
		},
		{
			name: "workspace_tuple_without_metadata",
			pkgs: []any{"a@workspace:packages/a"},
		},
		{
			name:     "non_string_dep_values_skipped",
			pkgs:     []any{"a@1.0.0", "", map[string]any{"dependencies": map[string]any{"b": "^1.0.0", "c": 1}}, "sha512-..."},
			wantDeps: map[string]string{"b": "^1.0.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, optional, peer := packageDependencies(tt.pkgs)
			if diff := cmp.Diff(tt.wantDeps, deps); diff != "" {
				t.Errorf("dependencies diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantOptional, optional); diff != "" {
				t.Errorf("optionalDependencies diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantPeer, peer); diff != "" {
				t.Errorf("peerDependencies diff (-want +got):\n%s", diff)
			}
		})
	}
}
