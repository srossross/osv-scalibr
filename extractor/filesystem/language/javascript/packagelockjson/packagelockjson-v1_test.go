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

func TestNPMLockExtractor_Extract_V1(t *testing.T) {
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
				Path: "testdata/empty.v1.json",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-wrappy-1",
					Name:       "wrappy",
					Version:    "1.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/one-package.v1.json"),
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
				Path: "testdata/one-package-dev.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-wrappy-1",
					Name:       "wrappy",
					Version:    "1.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/one-package-dev.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
			},
		},
		{
			Name: "two packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/two-packages.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-wrappy-2",
					Name:       "wrappy",
					Version:    "1.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/two-packages.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-supports-color-1",
					Name:       "supports-color",
					Version:    "5.5.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/two-packages.v1.json"),
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
				Path: "testdata/scoped-packages.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-wrappy-2",
					Name:       "wrappy",
					Version:    "1.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/scoped-packages.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@babel/code-frame-1",
					Name:       "@babel/code-frame",
					Version:    "7.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/scoped-packages.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "nested dependencies",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/nested-dependencies.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-postcss-2",
					Name:       "postcss",
					Version:    "6.0.23",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-3",
					Name:       "postcss",
					ParentIDs:  map[string]bool{"id-postcss-calc-1": true},
					Version:    "7.0.16",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-calc-1",
					Name:       "postcss-calc",
					Version:    "7.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-supports-color-5",
					Name:       "supports-color",
					ParentIDs:  map[string]bool{"id-postcss-3": true},
					Version:    "6.1.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-supports-color-4",
					Name:       "supports-color",
					ParentIDs:  map[string]bool{"id-postcss-2": true},
					Version:    "5.5.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "nested dependencies dup",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/nested-dependencies-dup.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-supports-color-37",
					Name:       "supports-color",
					Version:    "2.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-display-values-24",
					Name:       "postcss-normalize-display-values",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-timing-functions-28",
					Name:       "postcss-normalize-timing-functions",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-string-27",
					Name:       "postcss-normalize-string",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-whitespace-31",
					Name:       "postcss-normalize-whitespace",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-supports-color-39",
					Name:       "supports-color",
					Version:    "6.1.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-cssnano-preset-default-5",
					Name:       "cssnano-preset-default",
					ParentIDs:  map[string]bool{"id-cssnano-7": true},
					Version:    "4.0.7",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-merge-longhand-17",
					Name:       "postcss-merge-longhand",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.11",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-discard-overridden-15",
					Name:       "postcss-discard-overridden",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-reduce-transforms-34",
					Name:       "postcss-reduce-transforms",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-svgo-35",
					Name:       "postcss-svgo",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-ordered-values-32",
					Name:       "postcss-ordered-values",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.1.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-minify-selectors-22",
					Name:       "postcss-minify-selectors",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-babel-code-frame-3",
					Name:       "babel-code-frame",
					Version:    "6.26.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-css-declaration-sorter-4",
					Name:       "css-declaration-sorter",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-url-30",
					Name:       "postcss-normalize-url",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-minify-params-21",
					Name:       "postcss-minify-params",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-supports-color-38",
					Name:       "supports-color",
					Version:    "5.5.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-colormin-10",
					Name:       "postcss-colormin",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-autoprefixer-2",
					Name:       "autoprefixer",
					Version:    "9.5.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-charset-23",
					Name:       "postcss-normalize-charset",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-unique-selectors-36",
					Name:       "postcss-unique-selectors",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-reduce-initial-33",
					Name:       "postcss-reduce-initial",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-positions-25",
					Name:       "postcss-normalize-positions",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-discard-duplicates-13",
					Name:       "postcss-discard-duplicates",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-loader-16",
					Name:       "postcss-loader",
					Version:    "3.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-cssnano-7",
					Name:       "cssnano",
					Version:    "4.1.10",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-discard-empty-14",
					Name:       "postcss-discard-empty",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-repeat-style-26",
					Name:       "postcss-normalize-repeat-style",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-convert-values-11",
					Name:       "postcss-convert-values",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-friendly-errors-webpack-plugin-8",
					Name:       "friendly-errors-webpack-plugin",
					Version:    "1.7.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-@vue/component-compiler-utils-1",
					Name:       "@vue/component-compiler-utils",
					Version:    "2.6.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-merge-rules-18",
					Name:       "postcss-merge-rules",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-normalize-unicode-29",
					Name:       "postcss-normalize-unicode",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-minify-font-values-19",
					Name:       "postcss-minify-font-values",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-minify-gradients-20",
					Name:       "postcss-minify-gradients",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-cssnano-util-raw-cache-6",
					Name:       "cssnano-util-raw-cache",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-calc-9",
					Name:       "postcss-calc",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "7.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-postcss-discard-comments-12",
					Name:       "postcss-discard-comments",
					ParentIDs:  map[string]bool{"id-cssnano-preset-default-5": true},
					Version:    "4.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/nested-dependencies-dup.v1.json"),
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
				Path: "testdata/commits.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-@segment/analytics.js-integration-facebook-pixel-1",
					Name:     "@segment/analytics.js-integration-facebook-pixel",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "3b1bb80b302c2e552685dc8a029797ec832ea7c9",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-ansi-styles-2",
					Name:       "ansi-styles",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-babel-preset-php-3",
					Name:     "babel-preset-php",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "c5a7ba5e0ad98b8db1cb8ce105403dd4b768cced",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-is-number-1-4",
					Name:     "is-number-1",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
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
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "be5935f8d2595bcd97b05718ef1eeae08d812e10",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-2-7",
					Name:     "is-number-2",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-is-number-2-6",
					Name:     "is-number-2",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "82dcc8e914dabd9305ab9ae580709a7825e824f5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-is-number-3-9",
					Name:     "is-number-3",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
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
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "82ae8802978da40d7f1be5ad5943c9e550ab2c89",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-4-10",
					Name:     "is-number-4",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-5-11",
					Name:     "is-number-5",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:       "id-is-number-6-12",
					Name:     "is-number-6",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:         "id-postcss-calc-13",
					Name:       "postcss-calc",
					Version:    "7.0.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-raven-js-14",
					Name:     "raven-js",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "c2b377e7a254264fd4a1fe328e4e3cfc9e245570",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:       "id-slick-carousel-15",
					Name:     "slick-carousel",
					Version:  "",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPath("testdata/commits.v1.json"),
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
				Path: "testdata/files.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-lodash-1",
					Name:       "lodash",
					ParentIDs:  map[string]bool{"id-other_package-2": true},
					Version:    "1.3.1",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/files.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-other_package-2",
					Name:       "other_package",
					Version:    "",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/files.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "alias",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/alias.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-@babel/code-frame-1",
					Name:       "@babel/code-frame",
					Version:    "7.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/alias.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-string-width-3",
					Name:       "string-width",
					Version:    "4.2.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/alias.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-string-width-2",
					Name:       "string-width",
					Version:    "5.1.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/alias.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "optional package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/optional-package.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-wrappy-2",
					Name:       "wrappy",
					Version:    "1.0.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/optional-package.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev", "optional"},
					},
				},
				{
					ID:         "id-supports-color-1",
					Name:       "supports-color",
					Version:    "5.5.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/optional-package.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"optional"},
					},
				},
			},
		},
		{
			Name: "same package different groups",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/same-package-different-groups.v1.json",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-eslint-2",
					Name:       "eslint",
					Version:    "1.2.3",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/same-package-different-groups.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev"},
					},
				},
				{
					ID:         "id-table-3",
					Name:       "table",
					Version:    "1.0.0",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/same-package-different-groups.v1.json"),
					SourceCode: &extractor.SourceCodeIdentifier{},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
				{
					ID:         "id-ajv-1",
					Name:       "ajv",
					Version:    "5.5.2",
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPath("testdata/same-package-different-groups.v1.json"),
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
