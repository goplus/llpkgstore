package mappingtable

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/goplus/llpkgstore/internal/actions/llpkg"
	vrs "github.com/goplus/llpkgstore/internal/actions/versions"
	"github.com/goplus/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

const _releaseSource = "https://github.com/goplus/llpkg/releases/download/_mappingtable/llpkgstore.json"

// Versions is a mapping table implement for Github Action only.
// It's recommend to use another implement in llgo for common usage.
type Versions struct {
	metadata.MetadataMap

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
		log.Fatalf("version %s has already existed", elem)
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

	return readFromBytes(b, f.Name())
}

func FromRelease() (table *Versions, isCreated bool, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", _releaseSource, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// if specified release is not created yet, return a blank mapping table.
	if resp.StatusCode == http.StatusNotFound {
		table = readFromBytes(nil, "llpkgstore.json")
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	isCreated = true
	table = readFromBytes(b, "llpkgstore.json")
	return
}

func readFromBytes(b []byte, fileName string) *Versions {
	m := metadata.MetadataMap{}

	if len(b) > 0 {
		json.Unmarshal(b, &m)
	}

	return &Versions{
		MetadataMap: m,
		fileName:    fileName,
	}
}

// cVersions retrieves the version mappings for a specific C library.
// It returns a map where keys are C library versions and values are supported Go versions.
func (v *Versions) cVersions(clib llpkg.ClibName) map[metadata.CVersion][]metadata.GoVersion {
	versions := v.MetadataMap[clib.String()]
	if versions == nil {
		return nil
	}
	return versions.Versions
}

// CVersions returns all available versions of the specified C library.
// The versions are returned as semantic version strings.
func (v *Versions) CVersions(clib llpkg.ClibName) (ret []string) {
	versions := v.MetadataMap[clib.String()]
	if versions == nil {
		return
	}
	for version := range versions.Versions {
		ret = append(ret, vrs.ToSemVer(version))
	}
	return
}

// GoVersions lists all Go versions associated with the given C library.
func (v *Versions) GoVersions(clib llpkg.ClibName) (ret []string) {
	versions := v.MetadataMap[clib.String()]
	if versions == nil {
		return
	}
	for _, goversion := range versions.Versions {
		ret = append(ret, goversion...)
	}
	return
}

// LatestGoVersionForCVersion finds the latest Go version compatible with a specific C library version.
func (v *Versions) LatestGoVersionForCVersion(clib llpkg.ClibName, cver llpkg.ClibVersion) string {
	version := v.MetadataMap[clib.String()]
	if version == nil {
		return ""
	}
	goVersions := version.Versions[cver.String()]
	if len(goVersions) == 0 {
		return ""
	}
	semver.Sort(goVersions)
	return goVersions[len(goVersions)-1]
}

// SearchBySemVer looks up a C library version by its semantic version string.
func (v *Versions) SearchBySemVer(clib llpkg.ClibName, semver string) llpkg.ClibVersion {
	for version := range v.cVersions(clib) {
		if vrs.ToSemVer(version) == semver {
			return llpkg.ClibVersion(version)
		}
	}
	return ""
}

// LatestGoVersion retrieves the latest Go version associated with the specified C library.
// It aggregates all Go versions across all C library versions and returns the highest one based on semantic versioning.
func (v *Versions) LatestGoVersion(clib llpkg.ClibName) string {
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
func (v *Versions) Write(clib llpkg.ClibName, clibVersion llpkg.ClibVersion, mappedVersion string) {
	clibName := clib.String()
	cversion := clibVersion.String()

	clibVersions := v.MetadataMap[clibName]
	if clibVersions == nil {
		clibVersions = &metadata.Metadata{
			Versions: map[metadata.CVersion][]metadata.GoVersion{},
		}
		v.MetadataMap[clibName] = clibVersions
	}
	versions := clibVersions.Versions[cversion]

	versions = appendVersion(versions, mappedVersion)

	clibVersions.Versions[cversion] = versions
	// sync to disk
	b, _ := json.Marshal(&v.MetadataMap)

	os.WriteFile(v.fileName, []byte(b), 0644)
}

// String returns the JSON representation of the Versions metadata.
func (v *Versions) String() string {
	b, _ := json.Marshal(&v.MetadataMap)
	return string(b)
}
