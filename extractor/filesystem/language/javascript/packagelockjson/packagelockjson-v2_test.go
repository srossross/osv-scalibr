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

package packagelockjson_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/packagelockjson"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/inventory/location"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"
	"github.com/google/osv-scalibr/testing/testcollector"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

func TestNPMLockExtractor_Extract_V2(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "invalid json",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/not-json.txt",
			},
			WantErr: extracttest.ContainsErrStr{Str: "could not extract"},
		},
		{
			Name: "no packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/empty.v2.json",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-wrappy-1",
					Name:      "wrappy",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/one-package.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "one package dev",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package-dev.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-wrappy-1",
					Name:      "wrappy",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/one-package-dev.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
			},
		},
		{
			Name: "two packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/two-packages.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-wrappy-2",
					Name:      "wrappy",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/two-packages.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-supports-color-1",
					Name:      "supports-color",
					ParentIDs: map[string]bool{"root": true},
					Version:   "5.5.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/two-packages.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "scoped packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/scoped-packages.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-wrappy-2",
					Name:     "wrappy",
					Version:  "1.0.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/scoped-packages.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-@babel/code-frame-1",
					Name:     "@babel/code-frame",
					Version:  "7.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/scoped-packages.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "nested dependencies",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/nested-dependencies.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-postcss-2",
					Name:     "postcss",
					Version:  "6.0.23",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/nested-dependencies.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-postcss-3",
					Name:      "postcss",
					ParentIDs: map[string]bool{"id-postcss-calc-1": true},
					Version:   "7.0.16",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/nested-dependencies.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-postcss-calc-1",
					Name:     "postcss-calc",
					Version:  "7.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/nested-dependencies.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-supports-color-5",
					Name:      "supports-color",
					ParentIDs: map[string]bool{"id-postcss-3": true},
					Version:   "6.1.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/nested-dependencies.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-supports-color-4",
					Name:      "supports-color",
					ParentIDs: map[string]bool{"id-postcss-2": true},
					Version:   "5.5.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/nested-dependencies.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "nested dependencies dup",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/nested-dependencies-dup.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-supports-color-2",
					Name:     "supports-color",
					Version:  "6.1.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/nested-dependencies-dup.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-supports-color-1",
					Name:     "supports-color",
					Version:  "2.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/nested-dependencies-dup.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "commits",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/commits.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-@segment/analytics.js-integration-facebook-pixel-1",
					Name:      "@segment/analytics.js-integration-facebook-pixel",
					ParentIDs: map[string]bool{"root": true},
					Version:   "2.4.1",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "3b1bb80b302c2e552685dc8a029797ec832ea7c9",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-ansi-styles-2",
					Name:     "ansi-styles",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-babel-preset-php-3",
					Name:      "babel-preset-php",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.1.1",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "c5a7ba5e0ad98b8db1cb8ce105403dd4b768cced",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:        "id-is-number-1-4",
					Name:      "is-number-1",
					ParentIDs: map[string]bool{"root": true},
					Version:   "3.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-1-5",
					Name:     "is-number-1",
					Version:  "3.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "be5935f8d2595bcd97b05718ef1eeae08d812e10",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:        "id-is-number-2-7",
					Name:      "is-number-2",
					ParentIDs: map[string]bool{"root": true},
					Version:   "2.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-2-6",
					Name:     "is-number-2",
					Version:  "2.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "82dcc8e914dabd9305ab9ae580709a7825e824f5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:        "id-is-number-3-9",
					Name:      "is-number-3",
					ParentIDs: map[string]bool{"root": true},
					Version:   "2.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-3-8",
					Name:     "is-number-3",
					Version:  "3.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "82ae8802978da40d7f1be5ad5943c9e550ab2c89",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:        "id-is-number-4-10",
					Name:      "is-number-4",
					ParentIDs: map[string]bool{"root": true},
					Version:   "3.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:        "id-is-number-5-11",
					Name:      "is-number-5",
					ParentIDs: map[string]bool{"root": true},
					Version:   "3.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-postcss-calc-12",
					Name:     "postcss-calc",
					Version:  "7.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-raven-js-13",
					Name:      "raven-js",
					ParentIDs: map[string]bool{"root": true},
					Version:   "",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "c2b377e7a254264fd4a1fe328e4e3cfc9e245570",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-slick-carousel-14",
					Name:      "slick-carousel",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.7.1",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/commits.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "280b560161b751ba226d50c7db1e0a14a78c2de0",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
			},
		},
		{
			Name: "files",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/files.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-etag-3",
					Name:     "etag",
					Version:  "1.8.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/files.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-abbrev-1",
					Name:     "abbrev",
					Version:  "1.0.9",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/files.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-abbrev-2",
					Name:     "abbrev",
					Version:  "2.3.4",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/files.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
			},
		},
		{
			Name: "alias",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/alias.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-@babel/code-frame-1",
					Name:      "@babel/code-frame",
					ParentIDs: map[string]bool{"root": true},
					Version:   "7.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/alias.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-string-width-2",
					Name:      "string-width",
					ParentIDs: map[string]bool{"root": true},
					Version:   "4.2.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/alias.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-string-width-3",
					Name:      "string-width",
					ParentIDs: map[string]bool{"root": true},
					Version:   "5.1.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/alias.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "optional package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/optional-package.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-wrappy-2",
					Name:      "wrappy",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.0.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/optional-package.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"optional"},
					},
				},
				{
					ID:       "id-supports-color-1",
					Name:     "supports-color",
					Version:  "5.5.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/optional-package.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev", "optional"},
					},
				},
			},
		},
		{
			Name: "same package different groups",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/same-package-different-groups.v2.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-eslint-2",
					Name:     "eslint",
					Version:  "1.2.3",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/same-package-different-groups.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-table-3",
					Name:     "table",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/same-package-different-groups.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-ajv-1",
					Name:     "ajv",
					Version:  "5.5.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/same-package-different-groups.v2.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "bundled_dependencies_are_grouped",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/bundled-dependencies.v3.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-ansi-regex-1",
					Name:      "ansi-regex",
					ParentIDs: map[string]bool{"id-strip-ansi-3": true},
					Version:   "6.2.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/bundled-dependencies.v3.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"bundled"},
					},
				},
				{
					ID:        "id-semver-2",
					Name:      "semver",
					ParentIDs: map[string]bool{"root": true},
					Version:   "7.7.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/bundled-dependencies.v3.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"bundled", "dev"},
					},
				},
				{
					ID:        "id-strip-ansi-3",
					Name:      "strip-ansi",
					ParentIDs: map[string]bool{"root": true},
					Version:   "7.1.2",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/bundled-dependencies.v3.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "workspace symlinks resolve to the linked package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/workspace-link/package-lock.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-lodash-1",
					Name:      "lodash",
					ParentIDs: map[string]bool{"id-w-2": true},
					Version:   "4.17.21",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/workspace-link/package-lock.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:        "id-w-2",
					Name:      "w",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPath("testdata/workspace-link/package-lock.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			collector := testcollector.New()
			extractor.SetIDGenerator(&testIDGenerator{})
			t.Cleanup(func() { extractor.SetIDGenerator(&extractor.RandomIDGenerator{}) })

			extr, err := packagelockjson.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("packagelockjson.New: %v", err)
			}
			extr.(*packagelockjson.Extractor).Stats = collector

			scanInput := extracttest.GenerateScanInputMock(t, tt.InputConfig)
			defer extracttest.CloseTestScanInput(t, scanInput)

			got, err := extr.Extract(t.Context(), &scanInput)

			if diff := cmp.Diff(tt.WantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s.Extract(%q) error diff (-want +got):\n%s", extr.Name(), tt.InputConfig.Path, diff)
				return
			}

			wantInv := inventory.Inventory{Packages: tt.WantPackages}
			if diff := cmp.Diff(wantInv, got, cmpopts.SortSlices(extracttest.PackageCmpLess), cmpopts.IgnoreFields(location.File{}, "LineNumber")); diff != "" {
				t.Errorf("%s.Extract(%q) diff (-want +got):\n%s", extr.Name(), tt.InputConfig.Path, diff)
			}

			gotFileSizeMetric := collector.FileExtractedFileSize(tt.InputConfig.Path)
			if gotFileSizeMetric != scanInput.Info.Size() {
				t.Errorf("Extract(%s) recorded file size %v, want file size %v", tt.InputConfig.Path, gotFileSizeMetric, scanInput.Info.Size())
			}
		})
	}
}
