// Copyright 2022 by LunaSec (owned by Refinery Labs, Inc)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package inventory

import (
	"fmt"
	"log"
	"lunasec/lunatrace/inventory/syftmodel"
	"lunasec/lunatrace/pkg/constants"
	"sort"
	"strconv"

	"github.com/anchore/syft/syft/file"

	"github.com/anchore/syft/syft/artifact"

	"github.com/anchore/syft/syft/sbom"

	"github.com/anchore/syft/syft/distro"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/source"
)

func toSyftJsonFormatModel(s *sbom.SBOM) syftmodel.Document {
	src, err := toSourceModel(s.Source)
	if err != nil {
		log.Printf("unable to create syft-json source object: %+v", err)
	}

	return syftmodel.Document{
		Artifacts:             toPackageModels(s.Artifacts.PackageCatalog),
		ArtifactRelationships: toRelationshipModel(s.Relationships),
		Files:                 toFile(*s),
		Secrets:               toSecrets(s.Artifacts.Secrets),
		Source:                src,
		Distro:                toDistroModel(s.Artifacts.Distro),
		Descriptor:            toDescriptor(s.Descriptor),
		Schema: syftmodel.Schema{
			Version: constants.JSONSchemaVersion,
			URL:     fmt.Sprintf("https://raw.githubusercontent.com/anchore/syft/main/schema/json/schema-%s.json", constants.JSONSchemaVersion),
		},
	}
}

func toDescriptor(d sbom.Descriptor) syftmodel.Descriptor {
	return syftmodel.Descriptor{
		Name:          d.Name,
		Version:       d.Version,
		Configuration: d.Configuration,
	}
}

func toSecrets(data map[source.Coordinates][]file.SearchResult) []syftmodel.Secrets {
	results := make([]syftmodel.Secrets, 0)
	for coordinates, secrets := range data {
		results = append(results, syftmodel.Secrets{
			Location: coordinates,
			Secrets:  secrets,
		})
	}

	// sort by real path then virtual path to ensure the result is stable across multiple runs
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Location.RealPath < results[j].Location.RealPath
	})
	return results
}

func toFile(s sbom.SBOM) []syftmodel.File {
	results := make([]syftmodel.File, 0)
	artifacts := s.Artifacts

	for _, coordinates := range sbom.AllCoordinates(s) {
		var metadata *source.FileMetadata
		if metadataForLocation, exists := artifacts.FileMetadata[coordinates]; exists {
			metadata = &metadataForLocation
		}

		var digests []file.Digest
		if digestsForLocation, exists := artifacts.FileDigests[coordinates]; exists {
			digests = digestsForLocation
		}

		var classifications []file.Classification
		if classificationsForLocation, exists := artifacts.FileClassifications[coordinates]; exists {
			classifications = classificationsForLocation
		}

		var contents string
		if contentsForLocation, exists := artifacts.FileContents[coordinates]; exists {
			contents = contentsForLocation
		}

		results = append(results, syftmodel.File{
			ID:              string(coordinates.ID()),
			Location:        coordinates,
			Metadata:        toFileMetadataEntry(coordinates, metadata),
			Digests:         digests,
			Classifications: classifications,
			Contents:        contents,
		})
	}

	// sort by real path then virtual path to ensure the result is stable across multiple runs
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Location.RealPath < results[j].Location.RealPath
	})
	return results
}

func toFileMetadataEntry(coordinates source.Coordinates, metadata *source.FileMetadata) *syftmodel.FileMetadataEntry {
	if metadata == nil {
		return nil
	}

	mode, err := strconv.Atoi(fmt.Sprintf("%o", metadata.Mode))
	if err != nil {
		log.Printf("invalid mode found in file catalog @ location=%+v mode=%q: %+v", coordinates, metadata.Mode, err)
		mode = 0
	}

	return &syftmodel.FileMetadataEntry{
		Mode:            mode,
		Type:            metadata.Type,
		LinkDestination: metadata.LinkDestination,
		UserID:          metadata.UserID,
		GroupID:         metadata.GroupID,
		MIMEType:        metadata.MIMEType,
	}
}

func toPackageModels(catalog *pkg.Catalog) []syftmodel.Package {
	artifacts := make([]syftmodel.Package, 0)
	if catalog == nil {
		return artifacts
	}
	for _, p := range catalog.Sorted() {
		artifacts = append(artifacts, toPackageModel(p))
	}
	return artifacts
}

// toPackageModel crates a new Package from the given pkg.Package.
func toPackageModel(p pkg.Package) syftmodel.Package {
	var cpes = make([]string, len(p.CPEs))
	for i, c := range p.CPEs {
		cpes[i] = pkg.CPEString(c)
	}

	var licenses = make([]string, 0)
	if p.Licenses != nil {
		licenses = p.Licenses
	}

	var coordinates = make([]source.Coordinates, len(p.Locations))
	for i, l := range p.Locations {
		coordinates[i] = l.Coordinates
	}

	return syftmodel.Package{
		PackageBasicData: syftmodel.PackageBasicData{
			ID:        string(p.ID()),
			Name:      p.Name,
			Version:   p.Version,
			Type:      p.Type,
			FoundBy:   p.FoundBy,
			Locations: coordinates,
			Licenses:  licenses,
			Language:  p.Language,
			CPEs:      cpes,
			PURL:      p.PURL,
		},
		PackageCustomData: syftmodel.PackageCustomData{
			MetadataType: p.MetadataType,
			Metadata:     p.Metadata,
		},
	}
}

func toRelationshipModel(relationships []artifact.Relationship) []syftmodel.Relationship {
	result := make([]syftmodel.Relationship, len(relationships))
	for i, r := range relationships {
		result[i] = syftmodel.Relationship{
			Parent:   string(r.From.ID()),
			Child:    string(r.To.ID()),
			Type:     string(r.Type),
			Metadata: r.Data,
		}
	}
	return result
}

// toSourceModel creates a new source object to be represented into JSON.
func toSourceModel(src source.Metadata) (syftmodel.Source, error) {
	switch src.Scheme {
	case source.ImageScheme:
		return syftmodel.Source{
			Type:   "image",
			Target: src.ImageMetadata,
		}, nil
	case source.DirectoryScheme:
		return syftmodel.Source{
			Type:   "directory",
			Target: src.Path,
		}, nil
	case source.FileScheme:
		return syftmodel.Source{
			Type:   "file",
			Target: src.Path,
		}, nil
	default:
		return syftmodel.Source{}, fmt.Errorf("unsupported source: %q", src.Scheme)
	}
}

// toDistroModel creates a struct with the Linux distribution to be represented in JSON.
func toDistroModel(d *distro.Distro) syftmodel.Distro {
	if d == nil {
		return syftmodel.Distro{}
	}

	return syftmodel.Distro{
		Name:    d.Name(),
		Version: d.FullVersion(),
		IDLike:  d.IDLike,
	}
}
