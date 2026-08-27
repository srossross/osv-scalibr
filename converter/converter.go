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

// Package converter provides utility functions for converting SCALIBR's scan results to
// standardized inventory formats.
package converter

import (
	"maps"
	"slices"
	"time"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scalibr/converter/spdx"
	"github.com/google/osv-scalibr/extractor"
	cdxmeta "github.com/google/osv-scalibr/extractor/filesystem/sbom/cdx/metadata"
	spdxmeta "github.com/google/osv-scalibr/extractor/filesystem/sbom/spdx/metadata"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/inventory/location"
	"github.com/google/osv-scalibr/log"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/uuid"
	"github.com/spdx/tools-golang/spdx/v2/v2_3"
)

// ToPURL converts a SCALIBR package structure into a package URL.
func ToPURL(p *extractor.Package) *purl.PackageURL {
	return p.PURL()
}

// ToSPDX23 converts the SCALIBR scan results into an SPDX v2.3 document.
func ToSPDX23(i inventory.Inventory, c spdx.Config) *v2_3.Document {
	return spdx.ToSPDX23(i, c)
}

// ToSPDX30 converts the SCALIBR scan results into an SPDX v3.0.1 document.
func ToSPDX30(i inventory.Inventory, c spdx.Config3) *spdx.Document3 {
	return spdx.ToSPDX30(i, c)
}

// CDXConfig describes custom settings that should be applied to the generated CDX file.
type CDXConfig struct {
	ComponentName    string
	ComponentVersion string
	ComponentType    string
	Authors          []string
}

// ToCDX converts the SCALIBR scan results into a CycloneDX document.
func ToCDX(i inventory.Inventory, c CDXConfig) *cyclonedx.BOM {
	bom := cyclonedx.NewBOM()
	bom.Metadata = &cyclonedx.Metadata{
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Component: &cyclonedx.Component{
			Name:    c.ComponentName,
			Version: c.ComponentVersion,
			Type:    cyclonedx.ComponentType(c.ComponentType),
			BOMRef:  uuid.New().String(),
		},
		Tools: &cyclonedx.ToolsChoice{
			Components: &[]cyclonedx.Component{
				{
					Type: cyclonedx.ComponentTypeApplication,
					Name: "SCALIBR",
					ExternalReferences: &[]cyclonedx.ExternalReference{
						{
							URL:  "https://github.com/google/osv-scalibr",
							Type: cyclonedx.ERTypeWebsite,
						},
					},
				},
			},
		},
	}
	if len(c.Authors) > 0 {
		authors := make([]cyclonedx.OrganizationalContact, 0, len(c.Authors))
		for _, author := range c.Authors {
			authors = append(authors, cyclonedx.OrganizationalContact{
				Name: author,
			})
		}
		bom.Metadata.Authors = &authors
	}

	scalibrToBOMRef := make(map[string]string)
	pkgToBOMRef := make(map[*extractor.Package]string)
	comps := make([]cyclonedx.Component, 0, len(i.Packages))
	for _, pkg := range i.Packages {
		comp := cyclonedx.Component{
			BOMRef:  uuid.New().String(),
			Type:    cyclonedx.ComponentTypeLibrary,
			Name:    pkg.Name,
			Version: pkg.Version,
		}
		pkgToBOMRef[pkg] = comp.BOMRef
		if pkg.ID != "" {
			scalibrToBOMRef[pkg.ID] = comp.BOMRef
		}
		if pkg.Name != "" {
			scalibrToBOMRef[pkg.Name] = comp.BOMRef
		}
		if p := ToPURL(pkg); p != nil {
			comp.PackageURL = p.String()
		}
		if cpes := extractCPEs(pkg); len(cpes) > 0 {
			comp.CPE = cpes[0]
		}
		occ := []cyclonedx.EvidenceOccurrence{}
		occ = appendOccurrenceFromLocation(pkg.Location.Descriptor, occ)
		for _, r := range pkg.Location.Related {
			occ = appendOccurrenceFromLocation(&r, occ)
		}
		if len(occ) > 0 {
			comp.Evidence = &cyclonedx.Evidence{
				Occurrences: &occ,
			}
		}
		comps = append(comps, comp)
	}
	bom.Components = &comps

	deps := dependencies(i.Packages, pkgToBOMRef, bom.Metadata.Component.BOMRef, scalibrToBOMRef)
	bom.Dependencies = &deps

	return bom
}

// dependencies builds the CDX dependency graph from each package's ParentIDs.
// The metadata component is the graph root, and every component gets an entry so
// that an empty dependsOn means "no dependencies" rather than "unknown".
func dependencies(
	invPackages []*extractor.Package,
	pkgToBOMRef map[*extractor.Package]string,
	rootRef string,
	scalibrToBOMRef map[string]string,
) []cyclonedx.Dependency {
	children := make(map[string][]string)
	for _, pkg := range invPackages {
		ref, ok := pkgToBOMRef[pkg]
		if !ok {
			continue
		}
		for _, parentID := range slices.Sorted(maps.Keys(pkg.ParentIDs)) {
			parentRef := rootRef
			if parentID != "root" {
				var ok bool
				if parentRef, ok = scalibrToBOMRef[parentID]; !ok {
					log.Warnf("Parent package ID %q for package %v not found in inventory", parentID, pkg)
					continue
				}
			}
			children[parentRef] = append(children[parentRef], ref)
		}
	}

	deps := make([]cyclonedx.Dependency, 0, len(invPackages)+1)
	addDep := func(ref string) {
		c := children[ref]
		if c == nil {
			c = []string{}
		}
		deps = append(deps, cyclonedx.Dependency{Ref: ref, Dependencies: &c})
	}
	addDep(rootRef)
	for _, pkg := range invPackages {
		if ref, ok := pkgToBOMRef[pkg]; ok {
			addDep(ref)
		}
	}
	return deps
}

func extractCPEs(p *extractor.Package) []string {
	// Only the two SBOM package types support storing CPEs.
	if m, ok := p.Metadata.(*spdxmeta.Metadata); ok {
		return m.CPEs
	}
	if m, ok := p.Metadata.(*cdxmeta.Metadata); ok {
		return m.CPEs
	}
	return nil
}

func appendOccurrenceFromLocation(l *location.Location, occ []cyclonedx.EvidenceOccurrence) []cyclonedx.EvidenceOccurrence {
	if l != nil && l.File != nil {
		occ = append(occ, cyclonedx.EvidenceOccurrence{
			Location: l.File.Path,
		})
	}
	return occ
}
