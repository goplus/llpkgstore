package mappingtable

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"slices"

	"github.com/goplus/llpkgstore/internal/actions/version"
	"github.com/goplus/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

// Versions is a mapping table implement for Github Action only.
// It's recommend to use another implement in llgo for common usage.
type Versions struct {
	m metadata.MetadataMap

	fileName string
}

// appendVersion adds a new version to the slice while preventing duplicates.
// It panics if the element already exists in the array to enforce uniqueness constraints.
// Parameters:
//
//	arr: Slice of versions to modify
//	elem: Version to append
func appendVersion(arr []string, elem string) []string {
	if slices.Contains(arr, elem) {
		log.Panicf("version %s has already existed", elem)
	}
	return append(arr, elem)
}

// Read initializes a Versions struct by reading version mappings from a file.
// It creates the file if it doesn't exist and parses the JSON content into the MetadataMap.
// Parameters:
//
//	fileName: Path to the version mapping file
func Read(fileName string) *Versions {
	// read or create a file
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	m := metadata.MetadataMap{}

	if len(b) > 0 {
		json.Unmarshal(b, &m)
	}

	return &Versions{
		m:        m,
		fileName: f.Name(),
	}
}

// CVersions returns all available versions of the specified C library.
// The versions are returned as semantic version strings.
func (v *Versions) CVersions(clib string) (ret []string) {
	versions := v.m[clib]
	if versions == nil {
		return
	}
	for cversion := range versions.Versions {
		ret = append(ret, version.ToSemVer(cversion))
	}
	return
}

// GoVersions lists all Go versions associated with the given C library.
func (v *Versions) GoVersions(clib string) (ret []string) {
	versions := v.m[clib]
	if versions == nil {
		return
	}
	for _, goversion := range versions.Versions {
		ret = append(ret, goversion...)
	}
	return
}

// LatestGoVersionForCVersion finds the latest Go version compatible with a specific C library version.
func (v *Versions) LatestGoVersionForCVersion(clib, cver string) string {
	goVersions := v.goVersion(clib, cver)
	if len(goVersions) == 0 {
		return ""
	}
	semver.Sort(goVersions)
	return goVersions[len(goVersions)-1]
}

// SearchBySemVer looks up a C library version by its semantic version string.
func (v *Versions) SearchBySemVer(clib, semver string) string {
	for cversion := range v.cVersions(clib) {
		if version.ToSemVer(cversion) == semver {
			return cversion
		}
	}
	return ""
}

// LatestGoVersion retrieves the latest Go version associated with the specified C library.
// It aggregates all Go versions across all C library versions and returns the highest one based on semantic versioning.
func (v *Versions) LatestGoVersion(clib string) string {
	versions := v.GoVersions(clib)
	if len(versions) == 0 {
		return ""
	}
	semver.Sort(versions)
	return versions[len(versions)-1]
}

// Write records a new Go version mapping for a C library version and persists to file.
// Parameters:
//
//	clib: The C library name.
//	clibVersion: The specific version of the C library.
//	mappedVersion: The Go version to map with the C library version.
//
// It appends the Go version to the existing list for the C library version and saves the updated metadata.
func (v *Versions) Write(clib, clibVersion, mappedVersion string) {
	v.initCVersion(clib)

	versions := v.goVersion(clib, clibVersion)

	// TODO(ghl): rewrite llpkgstore.json
	v.m[clib].Versions[clibVersion] = appendVersion(versions, mappedVersion)

	// sync to disk
	b, _ := json.Marshal(&v.m)

	os.WriteFile(v.fileName, []byte(b), 0644)
}

// String returns the JSON representation of the Versions metadata.
func (v *Versions) String() string {
	b, _ := json.Marshal(&v.m)
	return string(b)
}

// initCVersion initializes the C library version entry in metadata if it doesn't exist
// Parameters:
//
//	clib: The name of the C library
func (v *Versions) initCVersion(clib string) {
	clibVersions := v.m[clib]
	if clibVersions == nil {
		clibVersions = &metadata.Metadata{
			Versions: map[metadata.CVersion][]metadata.GoVersion{},
		}
		v.m[clib] = clibVersions
	}
}

// goVersion retrieves Go versions associated with specific C library version
// Parameters:
//
//	clib: The C library name
//	clibVersion: The C library version
//
// Returns:
//
//	Slice of Go versions compatible with given C library version
func (v *Versions) goVersion(clib, clibVersion string) []metadata.GoVersion {
	clibVersions := v.m[clib]
	if clibVersions == nil {
		return nil
	}
	return slices.Clone(clibVersions.Versions[clibVersion])
}

// cVersions retrieves the version mappings for a specific C library.
// It returns a map where keys are C library versions and values are supported Go versions.
func (v *Versions) cVersions(clib string) map[metadata.CVersion][]metadata.GoVersion {
	versions := v.m[clib]
	if versions == nil {
		return nil
	}
	return versions.Versions
}
