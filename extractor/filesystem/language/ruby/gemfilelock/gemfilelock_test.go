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

package gemfilelock_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/language/ruby/gemfilelock"
	"github.com/google/osv-scalibr/extractor/filesystem/simplefileapi"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"

	cpb "github.com/google/osv-scalibr/binary/proto/config_go_proto"
)

func TestExtractor_FileRequired(t *testing.T) {
	tests := []struct {
		inputPath string
		want      bool
	}{
		{
			inputPath: "",
			want:      false,
		},
		{
			inputPath: "Gemfile.lock",
			want:      true,
		},
		{
			inputPath: "gems.locked",
			want:      true,
		},
		{
			inputPath: "path/to/my/Gemfile.lock",
			want:      true,
		},
		{
			inputPath: "path/to/my/gems.locked",
			want:      true,
		},
		{
			inputPath: "path/to/my/Gemfile.lock/file",
			want:      false,
		},
		{
			inputPath: "path/to/my/Gemfile.lock.file",
			want:      false,
		},
		{
			inputPath: "path.to.my.Gemfile.lock",
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.inputPath, func(t *testing.T) {
			e, err := gemfilelock.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("gemfilelock.New: %v", err)
			}
			got := e.FileRequired(simplefileapi.New(tt.inputPath, nil))
			if got != tt.want {
				t.Errorf("FileRequired(%s, FileInfo) got = %v, want %v", tt.inputPath, got, tt.want)
			}
		})
	}
}

// testIDGenerator produces IDs unique across duplicate package names, unlike
// mockidgenerator, and resettable per subtest, unlike SequentialIDGenerator.
type testIDGenerator struct{ counter int }

func (g *testIDGenerator) GenerateID(name string) (string, error) {
	g.counter++
	return fmt.Sprintf("id-%s-%d", name, g.counter), nil
}

func TestExtractor_Extract(t *testing.T) {
	tests := []extracttest.TestTableEntry{
		{
			Name: "no spec section",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/no-spec-section.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "no gem section",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/no-gem-section.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "no gems",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/no-gems.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "invalid spec",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/invalid.lock",
			},
			WantPackages: []*extractor.Package{},
		},
		{
			Name: "one gem",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/one-gem.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-ast-1",
					Name:      "ast",
					ParentIDs: map[string]bool{"root": true},
					Version:   "2.4.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/one-gem.lock", 4),
				},
			},
		},
		{
			Name: "trailing source section ",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/source-section-at-end.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-ast-1",
					Name:      "ast",
					ParentIDs: map[string]bool{"root": true},
					Version:   "2.4.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/source-section-at-end.lock", 16),
				},
			},
		},
		{
			Name: "some gems",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/some-gems.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-coderay-1",
					Name:      "coderay",
					ParentIDs: map[string]bool{"id-pry-3": true},
					Version:   "1.1.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/some-gems.lock", 4),
				},
				{
					ID:        "id-method_source-2",
					Name:      "method_source",
					ParentIDs: map[string]bool{"id-pry-3": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/some-gems.lock", 5),
				},
				{
					ID:        "id-pry-3",
					Name:      "pry",
					ParentIDs: map[string]bool{"root": true},
					Version:   "0.14.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/some-gems.lock", 6),
				},
			},
		},
		{
			Name: "multiple gems",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/multiple-gems.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-bundler-audit-1",
					Name:      "bundler-audit",
					ParentIDs: map[string]bool{"root": true},
					Version:   "0.9.0.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/multiple-gems.lock", 4),
				},
				{
					ID:        "id-coderay-2",
					Name:      "coderay",
					ParentIDs: map[string]bool{"id-pry-5": true},
					Version:   "1.1.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/multiple-gems.lock", 7),
				},
				{
					ID:        "id-dotenv-3",
					Name:      "dotenv",
					ParentIDs: map[string]bool{"root": true},
					Version:   "2.7.6",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/multiple-gems.lock", 8),
				},
				{
					ID:        "id-method_source-4",
					Name:      "method_source",
					ParentIDs: map[string]bool{"id-pry-5": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/multiple-gems.lock", 9),
				},
				{
					ID:        "id-pry-5",
					Name:      "pry",
					ParentIDs: map[string]bool{"root": true},
					Version:   "0.14.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/multiple-gems.lock", 10),
				},
				{
					ID:        "id-thor-6",
					Name:      "thor",
					ParentIDs: map[string]bool{"id-bundler-audit-1": true},
					Version:   "1.2.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/multiple-gems.lock", 13),
				},
			},
		},
		{
			Name: "rails",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/rails.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-actioncable-1",
					Name:      "actioncable",
					ParentIDs: map[string]bool{"id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 4),
				},
				{
					ID:        "id-actionmailbox-2",
					Name:      "actionmailbox",
					ParentIDs: map[string]bool{"id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 9),
				},
				{
					ID:        "id-actionmailer-3",
					Name:      "actionmailer",
					ParentIDs: map[string]bool{"id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 19),
				},
				{
					ID:        "id-actionpack-4",
					Name:      "actionpack",
					ParentIDs: map[string]bool{"id-actioncable-1": true, "id-actionmailbox-2": true, "id-actionmailer-3": true, "id-actiontext-5": true, "id-activestorage-10": true, "id-rails-35": true, "id-railties-38": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 29),
				},
				{
					ID:        "id-actiontext-5",
					Name:      "actiontext",
					ParentIDs: map[string]bool{"id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 36),
				},
				{
					ID:        "id-actionview-6",
					Name:      "actionview",
					ParentIDs: map[string]bool{"id-actionmailer-3": true, "id-actionpack-4": true, "id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 43),
				},
				{
					ID:        "id-activejob-7",
					Name:      "activejob",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actionmailer-3": true, "id-activestorage-10": true, "id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 49),
				},
				{
					ID:        "id-activemodel-8",
					Name:      "activemodel",
					ParentIDs: map[string]bool{"id-activerecord-9": true, "id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 52),
				},
				{
					ID:        "id-activerecord-9",
					Name:      "activerecord",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actiontext-5": true, "id-activestorage-10": true, "id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 54),
				},
				{
					ID:        "id-activestorage-10",
					Name:      "activestorage",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actiontext-5": true, "id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 57),
				},
				{
					ID:        "id-activesupport-11",
					Name:      "activesupport",
					ParentIDs: map[string]bool{"id-actioncable-1": true, "id-actionmailbox-2": true, "id-actionmailer-3": true, "id-actionpack-4": true, "id-actiontext-5": true, "id-actionview-6": true, "id-activejob-7": true, "id-activemodel-8": true, "id-activerecord-9": true, "id-activestorage-10": true, "id-globalid-17": true, "id-rails-35": true, "id-rails-dom-testing-36": true, "id-railties-38": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 64),
				},
				{
					ID:        "id-builder-12",
					Name:      "builder",
					ParentIDs: map[string]bool{"id-actionview-6": true},
					Version:   "3.2.4",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 69),
				},
				{
					ID:        "id-concurrent-ruby-13",
					Name:      "concurrent-ruby",
					ParentIDs: map[string]bool{"id-activesupport-11": true, "id-i18n-18": true, "id-tzinfo-43": true},
					Version:   "1.1.9",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 70),
				},
				{
					ID:        "id-crass-14",
					Name:      "crass",
					ParentIDs: map[string]bool{"id-loofah-20": true},
					Version:   "1.0.6",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 71),
				},
				{
					ID:        "id-digest-15",
					Name:      "digest",
					ParentIDs: map[string]bool{"id-net-imap-26": true, "id-net-pop-27": true, "id-net-smtp-29": true},
					Version:   "3.1.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 72),
				},
				{
					ID:        "id-erubi-16",
					Name:      "erubi",
					ParentIDs: map[string]bool{"id-actionview-6": true},
					Version:   "1.10.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 73),
				},
				{
					ID:        "id-globalid-17",
					Name:      "globalid",
					ParentIDs: map[string]bool{"id-actiontext-5": true, "id-activejob-7": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 74),
				},
				{
					ID:        "id-i18n-18",
					Name:      "i18n",
					ParentIDs: map[string]bool{"id-activesupport-11": true},
					Version:   "1.10.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 76),
				},
				{
					ID:        "id-io-wait-19",
					Name:      "io-wait",
					ParentIDs: map[string]bool{"id-net-protocol-28": true},
					Version:   "0.2.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 78),
				},
				{
					ID:        "id-loofah-20",
					Name:      "loofah",
					ParentIDs: map[string]bool{"id-rails-html-sanitizer-37": true},
					Version:   "2.14.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 79),
				},
				{
					ID:        "id-mail-21",
					Name:      "mail",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actionmailer-3": true},
					Version:   "2.7.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 82),
				},
				{
					ID:        "id-marcel-22",
					Name:      "marcel",
					ParentIDs: map[string]bool{"id-activestorage-10": true},
					Version:   "1.0.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 84),
				},
				{
					ID:        "id-method_source-23",
					Name:      "method_source",
					ParentIDs: map[string]bool{"id-railties-38": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 85),
				},
				{
					ID:        "id-mini_mime-24",
					Name:      "mini_mime",
					ParentIDs: map[string]bool{"id-activestorage-10": true, "id-mail-21": true},
					Version:   "1.1.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 86),
				},
				{
					ID:        "id-minitest-25",
					Name:      "minitest",
					ParentIDs: map[string]bool{"id-activesupport-11": true},
					Version:   "5.15.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 87),
				},
				{
					ID:        "id-net-imap-26",
					Name:      "net-imap",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actionmailer-3": true},
					Version:   "0.2.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 88),
				},
				{
					ID:        "id-net-pop-27",
					Name:      "net-pop",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actionmailer-3": true},
					Version:   "0.1.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 92),
				},
				{
					ID:        "id-net-protocol-28",
					Name:      "net-protocol",
					ParentIDs: map[string]bool{"id-net-imap-26": true, "id-net-pop-27": true, "id-net-smtp-29": true},
					Version:   "0.1.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 96),
				},
				{
					ID:        "id-net-smtp-29",
					Name:      "net-smtp",
					ParentIDs: map[string]bool{"id-actionmailbox-2": true, "id-actionmailer-3": true},
					Version:   "0.3.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 99),
				},
				{
					ID:        "id-nio4r-30",
					Name:      "nio4r",
					ParentIDs: map[string]bool{"id-actioncable-1": true},
					Version:   "2.5.8",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 103),
				},
				{
					ID:        "id-racc-32",
					Name:      "racc",
					ParentIDs: map[string]bool{"id-nokogiri-31": true},
					Version:   "1.6.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 106),
				},
				{
					ID:        "id-rack-33",
					Name:      "rack",
					ParentIDs: map[string]bool{"id-actionpack-4": true, "id-rack-test-34": true},
					Version:   "2.2.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 107),
				},
				{
					ID:        "id-rack-test-34",
					Name:      "rack-test",
					ParentIDs: map[string]bool{"id-actionpack-4": true},
					Version:   "1.1.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 108),
				},
				{
					ID:        "id-rails-35",
					Name:      "rails",
					ParentIDs: map[string]bool{"root": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 110),
				},
				{
					ID:        "id-rails-dom-testing-36",
					Name:      "rails-dom-testing",
					ParentIDs: map[string]bool{"id-actionmailer-3": true, "id-actionpack-4": true, "id-actionview-6": true},
					Version:   "2.0.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 124),
				},
				{
					ID:        "id-rails-html-sanitizer-37",
					Name:      "rails-html-sanitizer",
					ParentIDs: map[string]bool{"id-actionpack-4": true, "id-actionview-6": true},
					Version:   "1.4.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 127),
				},
				{
					ID:        "id-railties-38",
					Name:      "railties",
					ParentIDs: map[string]bool{"id-rails-35": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 129),
				},
				{
					ID:        "id-rake-39",
					Name:      "rake",
					ParentIDs: map[string]bool{"id-railties-38": true},
					Version:   "13.0.6",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 136),
				},
				{
					ID:        "id-strscan-40",
					Name:      "strscan",
					ParentIDs: map[string]bool{"id-net-imap-26": true},
					Version:   "3.0.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 137),
				},
				{
					ID:        "id-thor-41",
					Name:      "thor",
					ParentIDs: map[string]bool{"id-railties-38": true},
					Version:   "1.2.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 138),
				},
				{
					ID:        "id-timeout-42",
					Name:      "timeout",
					ParentIDs: map[string]bool{"id-net-pop-27": true, "id-net-protocol-28": true, "id-net-smtp-29": true},
					Version:   "0.2.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 139),
				},
				{
					ID:        "id-tzinfo-43",
					Name:      "tzinfo",
					ParentIDs: map[string]bool{"id-activesupport-11": true},
					Version:   "2.0.4",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 140),
				},
				{
					ID:        "id-websocket-driver-44",
					Name:      "websocket-driver",
					ParentIDs: map[string]bool{"id-actioncable-1": true},
					Version:   "0.7.5",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 142),
				},
				{
					ID:        "id-websocket-extensions-45",
					Name:      "websocket-extensions",
					ParentIDs: map[string]bool{"id-websocket-driver-44": true},
					Version:   "0.1.5",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 144),
				},
				{
					ID:        "id-zeitwerk-46",
					Name:      "zeitwerk",
					ParentIDs: map[string]bool{"id-railties-38": true},
					Version:   "2.5.4",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 145),
				},
				{
					ID:        "id-nokogiri-31",
					Name:      "nokogiri",
					ParentIDs: map[string]bool{"id-actiontext-5": true, "id-loofah-20": true, "id-rails-dom-testing-36": true},
					Version:   "1.13.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rails.lock", 104),
				},
			},
		},
		{
			Name: "rubocop",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/rubocop.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-ast-1",
					Name:      "ast",
					ParentIDs: map[string]bool{"id-parser-3": true},
					Version:   "2.4.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 4),
				},
				{
					ID:        "id-parallel-2",
					Name:      "parallel",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "1.21.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 5),
				},
				{
					ID:        "id-parser-3",
					Name:      "parser",
					ParentIDs: map[string]bool{"id-rubocop-7": true, "id-rubocop-ast-8": true},
					Version:   "3.1.1.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 6),
				},
				{
					ID:        "id-rainbow-4",
					Name:      "rainbow",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "3.1.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 8),
				},
				{
					ID:        "id-regexp_parser-5",
					Name:      "regexp_parser",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "2.2.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 9),
				},
				{
					ID:        "id-rexml-6",
					Name:      "rexml",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "3.2.5",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 10),
				},
				{
					ID:        "id-rubocop-7",
					Name:      "rubocop",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.25.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 11),
				},
				{
					ID:        "id-rubocop-ast-8",
					Name:      "rubocop-ast",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "1.16.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 20),
				},
				{
					ID:        "id-ruby-progressbar-9",
					Name:      "ruby-progressbar",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "1.11.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 22),
				},
				{
					ID:        "id-unicode-display_width-10",
					Name:      "unicode-display_width",
					ParentIDs: map[string]bool{"id-rubocop-7": true},
					Version:   "2.1.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/rubocop.lock", 23),
				},
			},
		},
		{
			Name: "has local gem",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/has-local-gem.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:        "id-backbone-on-rails-1",
					Name:      "backbone-on-rails",
					ParentIDs: map[string]bool{"root": true},
					Version:   "1.2.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 4),
				},
				{
					ID:        "id-actionpack-2",
					Name:      "actionpack",
					ParentIDs: map[string]bool{"id-railties-26": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 14),
				},
				{
					ID:        "id-actionview-3",
					Name:      "actionview",
					ParentIDs: map[string]bool{"id-actionpack-2": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 21),
				},
				{
					ID:        "id-activesupport-4",
					Name:      "activesupport",
					ParentIDs: map[string]bool{"id-actionpack-2": true, "id-actionview-3": true, "id-rails-dom-testing-24": true, "id-railties-26": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 27),
				},
				{
					ID:        "id-builder-5",
					Name:      "builder",
					ParentIDs: map[string]bool{"id-actionview-3": true},
					Version:   "3.2.4",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 32),
				},
				{
					ID:        "id-coffee-script-6",
					Name:      "coffee-script",
					ParentIDs: map[string]bool{"id-eco-10": true},
					Version:   "2.4.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 33),
				},
				{
					ID:        "id-coffee-script-source-7",
					Name:      "coffee-script-source",
					ParentIDs: map[string]bool{"id-coffee-script-6": true},
					Version:   "1.12.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 36),
				},
				{
					ID:        "id-concurrent-ruby-8",
					Name:      "concurrent-ruby",
					ParentIDs: map[string]bool{"id-activesupport-4": true, "id-i18n-15": true, "id-tzinfo-29": true},
					Version:   "1.1.9",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 37),
				},
				{
					ID:        "id-crass-9",
					Name:      "crass",
					ParentIDs: map[string]bool{"id-loofah-17": true},
					Version:   "1.0.6",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 38),
				},
				{
					ID:        "id-eco-10",
					Name:      "eco",
					ParentIDs: map[string]bool{"id-backbone-on-rails-1": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 39),
				},
				{
					ID:        "id-ejs-12",
					Name:      "ejs",
					ParentIDs: map[string]bool{"id-backbone-on-rails-1": true},
					Version:   "1.1.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 44),
				},
				{
					ID:        "id-erubi-13",
					Name:      "erubi",
					ParentIDs: map[string]bool{"id-actionview-3": true},
					Version:   "1.10.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 45),
				},
				{
					ID:        "id-execjs-14",
					Name:      "execjs",
					ParentIDs: map[string]bool{"id-coffee-script-6": true, "id-eco-10": true},
					Version:   "2.8.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 46),
				},
				{
					ID:        "id-i18n-15",
					Name:      "i18n",
					ParentIDs: map[string]bool{"id-activesupport-4": true},
					Version:   "1.10.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 47),
				},
				{
					ID:        "id-jquery-rails-16",
					Name:      "jquery-rails",
					ParentIDs: map[string]bool{"id-backbone-on-rails-1": true},
					Version:   "4.4.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 49),
				},
				{
					ID:        "id-loofah-17",
					Name:      "loofah",
					ParentIDs: map[string]bool{"id-rails-html-sanitizer-25": true},
					Version:   "2.14.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 53),
				},
				{
					ID:        "id-method_source-18",
					Name:      "method_source",
					ParentIDs: map[string]bool{"id-railties-26": true},
					Version:   "1.0.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 56),
				},
				{
					ID:        "id-minitest-19",
					Name:      "minitest",
					ParentIDs: map[string]bool{"id-activesupport-4": true},
					Version:   "5.15.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 57),
				},
				{
					ID:        "id-racc-21",
					Name:      "racc",
					ParentIDs: map[string]bool{"id-nokogiri-20": true},
					Version:   "1.6.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 60),
				},
				{
					ID:        "id-rack-22",
					Name:      "rack",
					ParentIDs: map[string]bool{"id-actionpack-2": true, "id-rack-test-23": true},
					Version:   "2.2.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 61),
				},
				{
					ID:        "id-rack-test-23",
					Name:      "rack-test",
					ParentIDs: map[string]bool{"id-actionpack-2": true},
					Version:   "1.1.0",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 62),
				},
				{
					ID:        "id-rails-dom-testing-24",
					Name:      "rails-dom-testing",
					ParentIDs: map[string]bool{"id-actionpack-2": true, "id-actionview-3": true, "id-jquery-rails-16": true},
					Version:   "2.0.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 64),
				},
				{
					ID:        "id-rails-html-sanitizer-25",
					Name:      "rails-html-sanitizer",
					ParentIDs: map[string]bool{"id-actionpack-2": true, "id-actionview-3": true},
					Version:   "1.4.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 67),
				},
				{
					ID:        "id-railties-26",
					Name:      "railties",
					ParentIDs: map[string]bool{"id-backbone-on-rails-1": true, "id-jquery-rails-16": true},
					Version:   "7.0.2.2",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 69),
				},
				{
					ID:        "id-rake-27",
					Name:      "rake",
					ParentIDs: map[string]bool{"id-railties-26": true},
					Version:   "13.0.6",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 76),
				},
				{
					ID:        "id-thor-28",
					Name:      "thor",
					ParentIDs: map[string]bool{"id-jquery-rails-16": true, "id-railties-26": true},
					Version:   "1.2.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 77),
				},
				{
					ID:        "id-tzinfo-29",
					Name:      "tzinfo",
					ParentIDs: map[string]bool{"id-activesupport-4": true},
					Version:   "2.0.4",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 78),
				},
				{
					ID:        "id-zeitwerk-30",
					Name:      "zeitwerk",
					ParentIDs: map[string]bool{"id-railties-26": true},
					Version:   "2.5.4",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 80),
				},
				{
					ID:        "id-nokogiri-20",
					Name:      "nokogiri",
					ParentIDs: map[string]bool{"id-loofah-17": true, "id-rails-dom-testing-24": true},
					Version:   "1.13.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 58),
				},
				{
					ID:        "id-eco-source-11",
					Name:      "eco-source",
					ParentIDs: map[string]bool{"id-eco-10": true},
					Version:   "1.1.0.rc.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-local-gem.lock", 43),
				},
			},
		},
		{
			Name: "has git gem",
			InputConfig: extracttest.ScanInputMockConfig{
				Path: "testdata/has-git-gem.lock",
			},
			WantPackages: []*extractor.Package{
				{
					ID:       "id-hanami-controller-1",
					Name:     "hanami-controller",
					Version:  "2.0.0.alpha1",
					PURLType: purl.TypeGem,
					Location: extractor.LocationFromPathAndLine("testdata/has-git-gem.lock", 6),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "027dbe2e56397b534e859fc283990cad1b6addd6",
					},
				},
				{
					ID:        "id-hanami-utils-2",
					Name:      "hanami-utils",
					ParentIDs: map[string]bool{"id-hanami-controller-1": true, "root": true},
					Version:   "2.0.0.alpha1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-git-gem.lock", 15),
					SourceCode: &extractor.SourceCodeIdentifier{
						Commit: "5904fc9a70683b8749aa2861257d0c8c01eae4aa",
					},
				},
				{
					ID:        "id-concurrent-ruby-3",
					Name:      "concurrent-ruby",
					ParentIDs: map[string]bool{"id-hanami-utils-2": true},
					Version:   "1.1.7",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-git-gem.lock", 22),
				},
				{
					ID:        "id-rack-4",
					Name:      "rack",
					ParentIDs: map[string]bool{"id-hanami-controller-1": true},
					Version:   "2.2.3",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-git-gem.lock", 23),
				},
				{
					ID:        "id-transproc-5",
					Name:      "transproc",
					ParentIDs: map[string]bool{"id-hanami-utils-2": true},
					Version:   "1.1.1",
					PURLType:  purl.TypeGem,
					Location:  extractor.LocationFromPathAndLine("testdata/has-git-gem.lock", 24),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			extractor.SetIDGenerator(&testIDGenerator{})
			t.Cleanup(func() { extractor.SetIDGenerator(&extractor.RandomIDGenerator{}) })

			extr, err := gemfilelock.New(&cpb.PluginConfig{})
			if err != nil {
				t.Fatalf("gemfilelock.New: %v", err)
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
