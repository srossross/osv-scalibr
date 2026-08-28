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

package pnpmlock_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/pnpmlock"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/extractor/filesystem/simplefileapi"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

// testIDGenerator produces IDs unique across duplicate package names, unlike
// mockidgenerator, and resettable per subtest, unlike SequentialIDGenerator.
type testIDGenerator struct{ counter int }

func (g *testIDGenerator) GenerateID(name string) (string, error) {
	g.counter++
	return fmt.Sprintf("id-%s-%d", name, g.counter), nil
}

func TestExtractor_FileRequired(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		want      bool
	}{
		{
			name:      "",
			inputPath: "",
			want:      false,
		},
		{
			name:      "",
			inputPath: "pnpm-lock.yaml",
			want:      true,
		},
		{
			name:      "",
			inputPath: "path/to/my/pnpm-lock.yaml",
			want:      true,
		},
		{
			name:      "",
			inputPath: "path/to/my/pnpm-lock.yaml/file",
			want:      false,
		},
		{
			name:      "",
			inputPath: "path/to/my/pnpm-lock.yaml.file",
			want:      false,
		},
		{
			name:      "",
			inputPath: "path.to.my.pnpm-lock.yaml",
			want:      false,
		},
		{
			name:      "",
			inputPath: "foo/node_modules/bar/pnpn-lock.yaml",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := pnpmlock.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("pnpmlock.New: %v", err)
			}
			got := e.FileRequired(simplefileapi.New(tt.inputPath, nil))
			if got != tt.want {
				t.Errorf("FileRequired(%s, FileInfo) got = %v, want %v", tt.inputPath, got, tt.want)
			}
		})
	}
}

func TestExtractor_Extract(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "invalid yaml",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/not-yaml.txt",
			},
			WantErr:      extracttest.ContainsErrStr{Str: "could not extract"},
			WantPackages: nil,
		},
		{
			Name: "invalid dep path",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/invalid-path.yaml",
			},
			WantErr: extracttest.ContainsErrStr{Str: "invalid dependency path"},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/invalid-path.yaml", 11),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "invalid dep paths (first error)",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/invalid-paths.yaml",
			},
			WantErr: extracttest.ContainsErrStr{Str: "invalid dependency path: invalidpath1"},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/invalid-paths.yaml", 17),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "invalid dep paths (second error)",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/invalid-paths.yaml",
			},
			WantErr: extracttest.ContainsErrStr{Str: "invalid dependency path: invalidpath2"},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/invalid-paths.yaml", 17),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "empty",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/empty.yaml",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "no packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/no-packages.yaml",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "missing packages key",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/missing-packages.yaml",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/one-package.yaml", 11),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "one package v6 lockfile",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package-v6-lockfile.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/one-package-v6-lockfile.yaml", 10),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "one package dev",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package-dev.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/one-package-dev.yaml", 11),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "scoped packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/scoped-packages.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@typescript-eslint/types-1",
					Name:       "@typescript-eslint/types",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/scoped-packages.yaml", 11),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "scoped packages v6 lockfile",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/scoped-packages-v6-lockfile.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@typescript-eslint/types-1",
					Name:       "@typescript-eslint/types",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.57.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/scoped-packages-v6-lockfile.yaml", 10),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "peer dependencies",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/peer-dependencies.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-jsx-1",
					Name:       "acorn-jsx",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies.yaml", 13),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-acorn-2",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"id-acorn-jsx-1": true, "root": true},
					Version:    "8.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies.yaml", 21),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "peer_dependencies_v6",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/peer-dependencies-v6.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-js-tokens-1",
					Name:       "js-tokens",
					ParentIDs:  map[string]bool{"id-loose-envify-2": true},
					Version:    "4.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-v6.yaml", 14),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-loose-envify-2",
					Name:       "loose-envify",
					ParentIDs:  map[string]bool{"id-react-4": true, "id-react-dom-3": true, "id-scheduler-5": true},
					Version:    "1.4.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-v6.yaml", 18),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-react-dom-3",
					Name:       "react-dom",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "18.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-v6.yaml", 25),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-react-4",
					Name:       "react",
					ParentIDs:  map[string]bool{"id-react-dom-3": true},
					Version:    "18.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-v6.yaml", 35),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-scheduler-5",
					Name:       "scheduler",
					ParentIDs:  map[string]bool{"id-react-dom-3": true},
					Version:    "0.23.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-v6.yaml", 42),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "peer dependencies advanced",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/peer-dependencies-advanced.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@typescript-eslint/eslint-plugin-1",
					Name:       "@typescript-eslint/eslint-plugin",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 17),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/parser-2",
					Name:       "@typescript-eslint/parser",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-1": true, "root": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 44),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/type-utils-3",
					Name:       "@typescript-eslint/type-utils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-1": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 64),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/types-4",
					Name:       "@typescript-eslint/types",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/parser-2": true, "id-@typescript-eslint/typescript-estree-5": true, "id-@typescript-eslint/utils-6": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 83),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/typescript-estree-5",
					Name:       "@typescript-eslint/typescript-estree",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/parser-2": true, "id-@typescript-eslint/utils-6": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 88),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/utils-6",
					Name:       "@typescript-eslint/utils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-1": true, "id-@typescript-eslint/type-utils-3": true},
					Version:    "5.13.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 109),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-eslint-utils-7",
					Name:       "eslint-utils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/utils-6": true, "id-eslint-8": true},
					Version:    "3.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 127),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-eslint-8",
					Name:       "eslint",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-1": true, "id-@typescript-eslint/parser-2": true, "id-@typescript-eslint/type-utils-3": true, "id-@typescript-eslint/utils-6": true, "id-eslint-utils-7": true, "root": true},
					Version:    "8.10.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 137),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-tsutils-9",
					Name:       "tsutils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-1": true, "id-@typescript-eslint/type-utils-3": true, "id-@typescript-eslint/typescript-estree-5": true},
					Version:    "3.21.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.yaml", 181),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "peer_dependencies_advanced_v6",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/peer-dependencies-advanced-v6.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-js-tokens-1",
					Name:       "js-tokens",
					ParentIDs:  map[string]bool{"id-loose-envify-2": true},
					Version:    "4.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-v6.yaml", 14),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-loose-envify-2",
					Name:       "loose-envify",
					ParentIDs:  map[string]bool{"id-react-4": true, "id-react-dom-3": true, "id-scheduler-5": true},
					Version:    "1.4.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-v6.yaml", 18),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-react-dom-3",
					Name:       "react-dom",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "18.3.0-canary-ab31a9ed2-20230824",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-v6.yaml", 25),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-react-4",
					Name:       "react",
					ParentIDs:  map[string]bool{"id-react-dom-3": true},
					Version:    "18.3.0-canary-ab31a9ed2-20230824",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-v6.yaml", 35),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-scheduler-5",
					Name:       "scheduler",
					ParentIDs:  map[string]bool{"id-react-dom-3": true},
					Version:    "0.24.0-canary-ab31a9ed2-20230824",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-v6.yaml", 42),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "peer_dependencies_advanced_rc_v6",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/peer-dependencies-advanced-rc-v6.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-js-tokens-1",
					Name:       "js-tokens",
					ParentIDs:  map[string]bool{"id-loose-envify-2": true},
					Version:    "4.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-rc-v6.yaml", 14),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-loose-envify-2",
					Name:       "loose-envify",
					ParentIDs:  map[string]bool{"id-react-4": true, "id-react-dom-3": true, "id-scheduler-5": true},
					Version:    "1.4.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-rc-v6.yaml", 18),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-react-dom-3",
					Name:       "react-dom",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "18.0.0-rc.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-rc-v6.yaml", 25),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-react-4",
					Name:       "react",
					ParentIDs:  map[string]bool{"id-react-dom-3": true},
					Version:    "18.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-rc-v6.yaml", 35),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-scheduler-5",
					Name:       "scheduler",
					ParentIDs:  map[string]bool{"id-react-dom-3": true},
					Version:    "0.21.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced-rc-v6.yaml", 42),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "multiple packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/multiple-packages.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-aws-sdk-1",
					Name:       "aws-sdk",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "2.1087.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 11),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-base64-js-2",
					Name:       "base64-js",
					ParentIDs:  map[string]bool{"id-buffer-3": true},
					Version:    "1.5.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 26),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-buffer-3",
					Name:       "buffer",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true},
					Version:    "4.9.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 30),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-events-4",
					Name:       "events",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true},
					Version:    "1.1.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 38),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-ieee754-5",
					Name:       "ieee754",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true, "id-buffer-3": true},
					Version:    "1.1.13",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 43),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-isarray-6",
					Name:       "isarray",
					ParentIDs:  map[string]bool{"id-buffer-3": true},
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 47),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-jmespath-7",
					Name:       "jmespath",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true},
					Version:    "0.16.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 51),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-punycode-8",
					Name:       "punycode",
					ParentIDs:  map[string]bool{"id-url-11": true},
					Version:    "1.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 56),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-querystring-9",
					Name:       "querystring",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true, "id-url-11": true},
					Version:    "0.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 60),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-sax-10",
					Name:       "sax",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true, "id-xml2js-13": true},
					Version:    "1.2.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 66),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-url-11",
					Name:       "url",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true},
					Version:    "0.10.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 70),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-uuid-12",
					Name:       "uuid",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true},
					Version:    "3.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 77),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-xml2js-13",
					Name:       "xml2js",
					ParentIDs:  map[string]bool{"id-aws-sdk-1": true},
					Version:    "0.4.19",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 83),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-xmlbuilder-14",
					Name:       "xmlbuilder",
					ParentIDs:  map[string]bool{"id-xml2js-13": true},
					Version:    "9.0.7",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-packages.yaml", 90),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "multiple versions",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/multiple-versions.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-uuid-1",
					Name:       "uuid",
					Version:    "3.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-versions.yaml", 13),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-uuid-2",
					Name:       "uuid",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "8.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-versions.yaml", 19),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-xmlbuilder-3",
					Name:       "xmlbuilder",
					Version:    "9.0.7",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-versions.yaml", 24),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "tarball",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/tarball.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@my-org/my-package-1",
					Name:       "@my-org/my-package",
					Version:    "3.2.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/tarball.yaml", 10),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
			},
		},
		{
			Name: "exotic",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/exotic.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-foo-2",
					Name:       "foo",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 10),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@foo/bar-1",
					Name:       "@foo/bar",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 11),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-foo-7",
					Name:       "foo",
					Version:    "1.1.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 12),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@foo/bar-6",
					Name:       "@foo/bar",
					Version:    "1.1.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 13),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-foo-4",
					Name:       "foo",
					Version:    "1.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 15),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-foo-5",
					Name:       "foo",
					Version:    "1.3.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 16),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-foo-3",
					Name:       "foo",
					Version:    "1.4.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/exotic.yaml", 17),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "commits",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/commits.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-my-bitbucket-package-1",
					Name:     "my-bitbucket-package",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.yaml", 14),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "6104ae42cd32c3d724036d3964678f197b2c9cdb",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-@my-scope/my-package-5",
					Name:     "@my-scope/my-package",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.yaml", 20),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "267087851ad5fac92a184749c27cd539e2fc862e",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-@my-scope/my-other-package-4",
					Name:     "@my-scope/my-other-package",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.yaml", 28),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "fbfc962ab51eb1d754749b68c064460221fbd689",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-faker-parser-2",
					Name:     "faker-parser",
					Version:  "0.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.yaml", 34),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "d2dc42a9351d4d89ec48c525e34f612b6d77993f",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-mocks-3",
					Name:     "mocks",
					Version:  "20.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.yaml", 42),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "590f321b4eb3f692bb211bd74e22947639a6f79d",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "files",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/files.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-my-file-package-5",
					Name:       "my-file-package",
					Version:    "0.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/files.yaml", 10),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-a-local-package-2",
					Name:       "a-local-package",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/files.yaml", 16),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-a-nested-local-package-3",
					Name:       "a-nested-local-package",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/files.yaml", 22),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-one-up-1",
					Name:       "one-up",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/files.yaml", 28),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-one-up-with-peer-4",
					Name:       "one-up-with-peer",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/files.yaml", 34),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			extractor.SetIDGenerator(&testIDGenerator{})
			t.Cleanup(func() { extractor.SetIDGenerator(&extractor.RandomIDGenerator{}) })

			extr, err := pnpmlock.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("pnpmlock.New: %v", err)
			}

			scanInput := extracttest.GenerateScanInputMock(t, tt.InputConfig)
			defer extracttest.CloseTestScanInput(t, scanInput)

			got, err := extr.Extract(t.Context(), &scanInput)

			if diff := cmp.Diff(tt.WantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s.Extract(%q) error diff (-want +got):\n%s", extr.Name(), tt.InputConfig.Path, diff)
				return
			}

			wantInv := inventory.Inventory{Packages: tt.WantPackages}
			if diff := cmp.Diff(wantInv, got, cmpopts.SortSlices(extracttest.PackageCmpLess)); diff != "" {
				t.Errorf("%s.Extract(%q) diff (-want +got):\n%s", extr.Name(), tt.InputConfig.Path, diff)
			}
		})
	}
}
