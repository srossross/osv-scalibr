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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/pnpmlock"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"
)

func TestExtractor_Extract_v9(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "no packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/no-packages.v9.yaml",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "8.11.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/one-package.v9.yaml", 17),
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
				Path: "testdata/one-package-dev.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-1",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "8.11.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/one-package-dev.v9.yaml", 17),
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
				Path: "testdata/scoped-packages.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@typescript-eslint/types-1",
					Name:       "@typescript-eslint/types",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.62.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/scoped-packages.v9.yaml", 17),
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
				Path: "testdata/peer-dependencies.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-acorn-jsx-1",
					Name:       "acorn-jsx",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies.v9.yaml", 17),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-acorn-2",
					Name:       "acorn",
					ParentIDs:  map[string]bool{"id-acorn-jsx-1": true},
					Version:    "8.11.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies.v9.yaml", 22),
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
				Path: "testdata/peer-dependencies-advanced.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@eslint-community/eslint-utils-1",
					Name:       "@eslint-community/eslint-utils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/utils-7": true, "id-eslint-9": true},
					Version:    "4.4.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 26),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@eslint/eslintrc-2",
					Name:       "@eslint/eslintrc",
					ParentIDs:  map[string]bool{"id-eslint-9": true},
					Version:    "2.1.4",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 32),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/eslint-plugin-3",
					Name:       "@typescript-eslint/eslint-plugin",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.62.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 36),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/parser-4",
					Name:       "@typescript-eslint/parser",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-3": true, "root": true},
					Version:    "5.62.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 47),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/type-utils-5",
					Name:       "@typescript-eslint/type-utils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-3": true},
					Version:    "5.62.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 57),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/typescript-estree-6",
					Name:       "@typescript-eslint/typescript-estree",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/parser-4": true, "id-@typescript-eslint/type-utils-5": true, "id-@typescript-eslint/utils-7": true},
					Version:    "5.62.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 67),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@typescript-eslint/utils-7",
					Name:       "@typescript-eslint/utils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-3": true, "id-@typescript-eslint/type-utils-5": true},
					Version:    "5.62.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 76),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-debug-8",
					Name:       "debug",
					ParentIDs:  map[string]bool{"id-@eslint/eslintrc-2": true, "id-@typescript-eslint/eslint-plugin-3": true, "id-@typescript-eslint/parser-4": true, "id-@typescript-eslint/type-utils-5": true, "id-@typescript-eslint/typescript-estree-6": true, "id-eslint-9": true},
					Version:    "4.3.4",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 82),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-eslint-9",
					Name:       "eslint",
					ParentIDs:  map[string]bool{"id-@eslint-community/eslint-utils-1": true, "id-@typescript-eslint/eslint-plugin-3": true, "id-@typescript-eslint/parser-4": true, "id-@typescript-eslint/type-utils-5": true, "id-@typescript-eslint/utils-7": true, "root": true},
					Version:    "8.57.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 91),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-has-flag-10",
					Name:       "has-flag",
					ParentIDs:  map[string]bool{"id-supports-color-11": true},
					Version:    "4.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 96),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-supports-color-11",
					Name:       "supports-color",
					Version:    "7.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 100),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-tsutils-12",
					Name:       "tsutils",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-3": true, "id-@typescript-eslint/type-utils-5": true, "id-@typescript-eslint/typescript-estree-6": true},
					Version:    "3.21.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 104),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-typescript-13",
					Name:       "typescript",
					ParentIDs:  map[string]bool{"id-@typescript-eslint/eslint-plugin-3": true, "id-@typescript-eslint/parser-4": true, "id-@typescript-eslint/type-utils-5": true, "id-@typescript-eslint/typescript-estree-6": true, "id-tsutils-12": true, "root": true},
					Version:    "4.9.5",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-dependencies-advanced.v9.yaml", 110),
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
				Path: "testdata/multiple-versions.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-uuid-1",
					Name:       "uuid",
					Version:    "8.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-versions.v9.yaml", 20),
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
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-versions.v9.yaml", 24),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-xmlbuilder-3",
					Name:       "xmlbuilder",
					Version:    "11.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/multiple-versions.v9.yaml", 28),
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
				Path: "testdata/commits.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-ansi-regex-1",
					Name:      "ansi-regex",
					ParentIDs: map[string]bool{"root": true},
					Version:   "6.0.1",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPathAndLine("testdata/commits.v9.yaml", 20),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "02fa893d619d3da85411acc8fd4e2eea0e95a9d9",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-is-number-2",
					Name:      "is-number",
					ParentIDs: map[string]bool{"root": true},
					Version:   "7.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPathAndLine("testdata/commits.v9.yaml", 25),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "98e8ff1da1a89f93d1397a24d7413ed15421c139",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "mixed groups",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/mixed-groups.v9.yaml",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-ansi-regex-1",
					Name:       "ansi-regex",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "5.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/mixed-groups.v9.yaml", 25),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-uuid-3",
					Name:       "uuid",
					Version:    "8.3.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/mixed-groups.v9.yaml", 33),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-is-number-2",
					Name:       "is-number",
					ParentIDs:  map[string]bool{"root": true},
					Version:    "7.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/mixed-groups.v9.yaml", 29),
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

			extr := pnpmlock.Extractor{}

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
