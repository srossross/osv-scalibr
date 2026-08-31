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

package yarnlock_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/javascript/yarnlock"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"
)

// pkgV1 builds an expected package from testdata/nested-v1/yarn.lock.
func pkgV1(t *testing.T, id, name, version string, line int, parents ...string) *extractor.Package {
	t.Helper()

	ids := map[string]bool{}
	for _, p := range parents {
		ids[p] = true
	}
	return &extractor.Package{
		ID:         id,
		Name:       name,
		Version:    version,
		ParentIDs:  ids,
		PURLType:   purl.TypeNPM,
		Location:   extractor.LocationFromPathAndLine("testdata/nested-v1/yarn.lock", line),
		SourceCode: &extractor.SourceCodeIdentifier{Commit: ""},
	}
}

func TestExtractor_Extract_v1(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "no packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/empty.v1.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-balanced-match-1",
					Name:     "balanced-match",
					Version:  "1.0.2",
					Location: extractor.LocationFromPathAndLine("testdata/one-package.v1.lock", 5),
					PURLType: purl.TypeNPM,
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "package with no version in header",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/no-version.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-balanced-match-1",
					Name:     "balanced-match",
					Version:  "1.0.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/no-version.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "two packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/two-packages.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-concat-stream-2",
					Name:     "concat-stream",
					Version:  "1.6.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/two-packages.v1.lock", 10),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-concat-map-1",
					Name:     "concat-map",
					Version:  "0.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/two-packages.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "with quotes",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/with-quotes.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-concat-stream-2",
					Name:     "concat-stream",
					Version:  "1.6.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-quotes.v1.lock", 10),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-concat-map-1",
					Name:     "concat-map",
					Version:  "0.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-quotes.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "multiple versions",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/multiple-versions.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-define-properties-1",
					Name:     "define-properties",
					Version:  "1.1.3",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/multiple-versions.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-define-property-2",
					Name:     "define-property",
					Version:  "0.2.5",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/multiple-versions.v1.lock", 12),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-define-property-3",
					Name:     "define-property",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/multiple-versions.v1.lock", 19),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-define-property-4",
					Name:     "define-property",
					Version:  "2.0.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/multiple-versions.v1.lock", 26),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "multiple constraints",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/multiple-constraints.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-@babel/code-frame-2",
					Name:     "@babel/code-frame",
					Version:  "7.12.13",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/multiple-constraints.v1.lock", 10),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-domelementtype-1",
					Name:     "domelementtype",
					Version:  "1.3.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/multiple-constraints.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "scoped packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/scoped-packages.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-@babel/code-frame-1",
					Name:     "@babel/code-frame",
					Version:  "7.12.11",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/scoped-packages.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-@babel/compat-data-2",
					Name:     "@babel/compat-data",
					Version:  "7.14.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/scoped-packages.v1.lock", 12),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "with prerelease",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/with-prerelease.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-css-tree-1",
					Name:     "css-tree",
					Version:  "1.0.0-alpha.37",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-prerelease.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-gensync-2",
					Name:     "gensync",
					Version:  "1.0.0-beta.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-prerelease.v1.lock", 13),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-node-fetch-3",
					Name:     "node-fetch",
					Version:  "3.0.0-beta.9",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-prerelease.v1.lock", 18),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-resolve-4",
					Name:     "resolve",
					Version:  "1.20.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-prerelease.v1.lock", 26),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-resolve-5",
					Name:     "resolve",
					Version:  "2.0.0-next.3",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-prerelease.v1.lock", 34),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "with build string",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/with-build-string.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-domino-1",
					Name:     "domino",
					Version:  "2.1.6+git",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-build-string.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-tslib-2",
					Name:     "tslib",
					Version:  "2.6.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-build-string.v1.lock", 10),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "commits",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/commits.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-mine1-1",
					Name:     "mine1",
					Version:  "1.0.0-alpha.37",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "0a2d2506c1fe299691fc5db53a2097db3bd615bc",
					},
				},
				{
					ID:       "id-mine2-2",
					Name:     "mine2",
					Version:  "0.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 11),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "0a2d2506c1fe299691fc5db53a2097db3bd615bc",
					},
				},
				{
					ID:       "id-mine3-3",
					Name:     "mine3",
					Version:  "1.2.3",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 17),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "094e581aaf927d010e4b61d706ba584551dac502",
					},
				},
				{
					ID:       "id-mine4-4",
					Name:     "mine4",
					Version:  "0.0.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 21),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "aa3bdfcb1d845c79f14abb66f60d35b8a3ee5998",
					},
				},
				{
					ID:       "id-mine4-5",
					Name:     "mine4",
					Version:  "0.0.4",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 25),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "aa3bdfcb1d845c79f14abb66f60d35b8a3ee5998",
					},
				},
				{
					ID:       "id-my-package-6",
					Name:     "my-package",
					Version:  "1.8.3",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 29),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "b3bd3f1b3dad036e671251f5258beaae398f983a",
					},
				},
				{
					ID:       "id-@bower_components/angular-animate-7",
					Name:     "@bower_components/angular-animate",
					Version:  "1.4.14",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 33),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "e7f778fc054a086ba3326d898a00fa1bc78650a8",
					},
				},
				{
					ID:       "id-@bower_components/alertify-8",
					Name:     "@bower_components/alertify",
					Version:  "0.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 37),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "e7b6c46d76604d297c389d830817b611c9a8f17c",
					},
				},
				{
					ID:       "id-minimist-9",
					Name:     "minimist",
					Version:  "0.0.8",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 41),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "3754568bfd43a841d2d72d7fb54598635aea8fa4",
					},
				},
				{
					ID:       "id-bats-assert-10",
					Name:     "bats-assert",
					Version:  "2.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 45),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "4bdd58d3fbcdce3209033d44d884e87add1d8405",
					},
				},
				{
					ID:       "id-bats-support-11",
					Name:     "bats-support",
					Version:  "0.3.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 49),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "d140a65044b2d6810381935ae7f0c94c7023c8c3",
					},
				},
				{
					ID:       "id-bats-12",
					Name:     "bats",
					Version:  "1.5.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 53),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "172580d2ce19ee33780b5f1df817bbddced43789",
					},
				},
				{
					ID:       "id-vue-13",
					Name:     "vue",
					Version:  "2.6.12",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 57),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "bb253db0b3e17124b6d1fe93fbf2db35470a1347",
					},
				},
				{
					ID:       "id-kit-14",
					Name:     "kit",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 61),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "5b6830c0252eb73c6024d40a8ff5106d3023a2a6",
					},
				},
				{
					ID:       "id-casadistance-15",
					Name:     "casadistance",
					Version:  "1.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 68),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "f0308391f0c50104182bfb2332a53e4e523a4603",
					},
				},
				{
					ID:       "id-babel-preset-php-16",
					Name:     "babel-preset-php",
					Version:  "1.1.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 72),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "c5a7ba5e0ad98b8db1cb8ce105403dd4b768cced",
					},
				},
				{
					ID:       "id-is-number-17",
					Name:     "is-number",
					Version:  "2.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 78),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
					},
				},
				{
					ID:       "id-is-number-18",
					Name:     "is-number",
					Version:  "5.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/commits.v1.lock", 82),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "af885e2e890b9ef0875edd2b117305119ee5bdc5",
					},
				},
			},
		},
		{
			Name: "files",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/files.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-etag-1",
					Name:      "etag",
					ParentIDs: map[string]bool{"id-other_package-4": true},
					Version:   "1.8.1",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPathAndLine("testdata/files.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-filedep-2",
					Name:     "filedep",
					Version:  "1.2.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/files.v1.lock", 9),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:        "id-lodash-3",
					Name:      "lodash",
					ParentIDs: map[string]bool{"id-other_package-4": true},
					Version:   "1.3.1",
					PURLType:  purl.TypeNPM,
					Location:  extractor.LocationFromPathAndLine("testdata/files.v1.lock", 12),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-other_package-4",
					Name:     "other_package",
					Version:  "0.0.2",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/files.v1.lock", 16),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-sprintf-js-5",
					Name:     "sprintf-js",
					Version:  "0.0.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/files.v1.lock", 24),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-etag-6",
					Name:     "etag",
					Version:  "1.8.0",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/files.v1.lock", 27),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "with aliases",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/with-aliases.v1.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-@babel/helper-validator-identifier-3",
					Name:     "@babel/helper-validator-identifier",
					Version:  "7.22.20",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-aliases.v1.lock", 15),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-ansi-regex-2",
					Name:     "ansi-regex",
					Version:  "6.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-aliases.v1.lock", 10),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
				{
					ID:       "id-ansi-regex-1",
					Name:     "ansi-regex",
					Version:  "5.0.1",
					PURLType: purl.TypeNPM,
					Location: extractor.LocationFromPathAndLine("testdata/with-aliases.v1.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "",
					},
				},
			},
		},
		{
			Name: "nested dependencies with a sibling package.json",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/nested-v1/yarn.lock",
			},
			WantPackages: []*extractor.Package{
				pkgV1(t, "id-ansi-styles-1", "ansi-styles", "4.3.0", 5, "id-chalk-2"),
				pkgV1(t, "id-chalk-2", "chalk", "4.1.2", 11, "root"),
				pkgV1(t, "id-color-convert-3", "color-convert", "2.0.1", 18, "id-ansi-styles-1"),
				pkgV1(t, "id-color-name-4", "color-name", "1.1.4", 24, "id-color-convert-3"),
				pkgV1(t, "id-has-flag-5", "has-flag", "3.0.0", 28, "id-supports-color-7"),
				pkgV1(t, "id-has-flag-6", "has-flag", "4.0.0", 32, "id-supports-color-8"),
				pkgV1(t, "id-supports-color-7", "supports-color", "5.5.0", 36, "root"),
				pkgV1(t, "id-supports-color-8", "supports-color", "7.2.0", 42, "id-chalk-2"),
			},
		},
		{
			Name: "peer dependencies",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/peer-deps/yarn.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:         "id-a-1",
					Name:       "a",
					Version:    "1.0.0",
					ParentIDs:  map[string]bool{"root": true},
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-deps/yarn.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{Commit: ""},
				},
				{
					ID:         "id-react-2",
					Name:       "react",
					Version:    "17.0.2",
					ParentIDs:  map[string]bool{"root": true, "id-a-1": true},
					PURLType:   purl.TypeNPM,
					Location:   extractor.LocationFromPathAndLine("testdata/peer-deps/yarn.lock", 11),
					SourceCode: &extractor.SourceCodeIdentifier{Commit: ""},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			extractor.SetIDGenerator(&testIDGenerator{})
			t.Cleanup(func() { extractor.SetIDGenerator(&extractor.RandomIDGenerator{}) })

			extr := yarnlock.Extractor{}

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
