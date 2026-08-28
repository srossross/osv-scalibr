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

package uvlock_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/python/uvlock"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/extractor/filesystem/simplefileapi"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

func pkg(t *testing.T, id string, name string, version string, path string, line int, parents ...string) *extractor.Package {
	t.Helper()

	return &extractor.Package{
		ID:        id,
		Name:      name,
		Version:   version,
		ParentIDs: parentIDs(parents),
		PURLType:  purl.TypePyPi,
		Location:  extractor.LocationFromPathAndLine(path, line),
		Metadata: &osv.DepGroupMetadata{
			DepGroupVals: []string{},
		},
	}
}

func parentIDs(parents []string) map[string]bool {
	if len(parents) == 0 {
		return nil
	}
	m := make(map[string]bool, len(parents))
	for _, p := range parents {
		m[p] = true
	}
	return m
}

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
			inputPath: "uv.lock",
			want:      true,
		},
		{
			name:      "",
			inputPath: "path/to/my/uv.lock",
			want:      true,
		},
		{
			name:      "",
			inputPath: "path/to/my/uv.lock/file",
			want:      false,
		},
		{
			name:      "",
			inputPath: "path/to/my/uv.lock.file",
			want:      false,
		},
		{
			name:      "",
			inputPath: "path.to.my.uv.lock",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := uvlock.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("uvlock.New: %v", err)
			}
			got := e.FileRequired(simplefileapi.New(tt.inputPath, nil))
			if got != tt.want {
				t.Errorf("FileRequired(%q, FileInfo) got = %v, want %v", tt.inputPath, got, tt.want)
			}
		})
	}
}

func TestExtractor_Extract(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "invalid toml",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/not-toml.txt",
			},
			WantErr:      extracttest.ContainsErrStr{Str: "could not extract"},
			WantPackages: nil,
		},
		{
			Name: "empty file",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/empty.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "no dependencies",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/empty.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "no packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/empty.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one package",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-package.lock",
			},
			WantPackages: []*extractor.Package{
				pkg(t, "id-emoji-1", "emoji", "2.14.0", "testdata/one-package.lock", 5, "root"),
			},
		},
		{
			Name: "two packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/two-packages.lock",
			},
			WantPackages: []*extractor.Package{
				pkg(t, "id-emoji-1", "emoji", "2.14.0", "testdata/two-packages.lock", 5, "root"),
				pkg(t, "id-protobuf-2", "protobuf", "4.25.5", "testdata/two-packages.lock", 14, "root"),
			},
		},
		{
			Name: "source git",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/source-git.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-ruff-1",
					Name:      "ruff",
					Version:   "0.8.1",
					ParentIDs: map[string]bool{"root": true},
					PURLType:  purl.TypePyPi,
					Location:  extractor.LocationFromPathAndLine("testdata/source-git.lock", 5),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "84748be16341b76e073d117329f7f5f4ee2941ad",
					},
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{},
					},
				},
			},
		},
		{
			Name: "grouped packages",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/grouped-packages.lock",
			},
			WantPackages: []*extractor.Package{
				pkg(t, "id-emoji-4", "emoji", "2.14.0", "testdata/grouped-packages.lock", 60, "root"),
				{
					ID:        "id-click-2",
					Name:      "click",
					Version:   "8.1.7",
					ParentIDs: map[string]bool{"id-black-1": true, "root": true},
					PURLType:  purl.TypePyPi,
					Location:  extractor.LocationFromPathAndLine("testdata/grouped-packages.lock", 39),
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"cli"},
					},
				},
				pkg(t, "id-colorama-3", "colorama", "0.4.6", "testdata/grouped-packages.lock", 51, "id-click-2"),
				{
					ID:        "id-black-1",
					Name:      "black",
					Version:   "24.10.0",
					ParentIDs: map[string]bool{"root": true},
					PURLType:  purl.TypePyPi,
					Location:  extractor.LocationFromPathAndLine("testdata/grouped-packages.lock", 5),
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"dev", "test"},
					},
				},
				{
					ID:        "id-flake8-5",
					Name:      "flake8",
					Version:   "7.1.1",
					ParentIDs: map[string]bool{"root": true},
					PURLType:  purl.TypePyPi,
					Location:  extractor.LocationFromPathAndLine("testdata/grouped-packages.lock", 69),
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"test"},
					},
				},
				pkg(t, "id-mccabe-6", "mccabe", "0.7.0", "testdata/grouped-packages.lock", 83, "id-flake8-5"),
				pkg(t, "id-mypy-extensions-7", "mypy-extensions", "1.0.0", "testdata/grouped-packages.lock", 92, "id-black-1"),
				pkg(t, "id-packaging-8", "packaging", "24.2", "testdata/grouped-packages.lock", 101, "id-black-1"),
				pkg(t, "id-pathspec-9", "pathspec", "0.12.1", "testdata/grouped-packages.lock", 110, "id-black-1"),
				pkg(t, "id-platformdirs-10", "platformdirs", "4.3.6", "testdata/grouped-packages.lock", 119, "id-black-1"),
				pkg(t, "id-pycodestyle-11", "pycodestyle", "2.12.1", "testdata/grouped-packages.lock", 128, "id-flake8-5"),
				pkg(t, "id-pyflakes-12", "pyflakes", "3.2.0", "testdata/grouped-packages.lock", 137, "id-flake8-5"),
				pkg(t, "id-tomli-13", "tomli", "2.2.1", "testdata/grouped-packages.lock", 146, "id-black-1"),
				pkg(t, "id-typing-extensions-14", "typing-extensions", "4.12.2", "testdata/grouped-packages.lock", 185, "id-black-1"),
			},
		},
		{
			Name: "editable project",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/editable-project.lock",
			},
			WantPackages: []*extractor.Package{
				pkg(t, "id-certifi-1", "certifi", "2024.12.14", "testdata/editable-project.lock", 5, "id-requests-5"),
				{
					ID:        "id-emoji-2",
					Name:      "emoji",
					Version:   "2.14.0",
					ParentIDs: map[string]bool{"root": true},
					PURLType:  purl.TypePyPi,
					Location:  extractor.LocationFromPathAndLine("testdata/editable-project.lock", 14),
					Metadata: &osv.DepGroupMetadata{
						DepGroupVals: []string{"cli"},
					},
				},
				pkg(t, "id-iniconfig-3", "iniconfig", "2.0.0", "testdata/editable-project.lock", 23, "id-pytest-4"),
				pkg(t, "id-pytest-4", "pytest", "8.3.4", "testdata/editable-project.lock", 32, "root"),
				pkg(t, "id-requests-5", "requests", "2.32.3", "testdata/editable-project.lock", 44, "root"),
			},
		},
		{
			Name: "names outside package block",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/names-outside-package-block.lock",
			},
			WantPackages: []*extractor.Package{
				pkg(t, "id-first-pkg-1", "first-pkg", "1.0.0", "testdata/names-outside-package-block.lock", 5),
				pkg(t, "id-second-pkg-2", "second-pkg", "2.0.0", "testdata/names-outside-package-block.lock", 18),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			extractor.SetIDGenerator(&testIDGenerator{})
			t.Cleanup(func() { extractor.SetIDGenerator(&extractor.RandomIDGenerator{}) })

			extr, err := uvlock.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("uvlock.New: %v", err)
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
