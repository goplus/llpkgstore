package versions

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"slices"

	"github.com/goplus/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

// Versions is a mapping table implement for Github Action only.
// It's recommend to use another implement in llgo for common usage.
type Versions struct {
	metadata.MetadataMap

	fileName    string
	cVerToGoVer map[string]*cVerMap
}

// appendVersion appends a version to an array, panic if the specified version has already existed.
func appendVersion(arr []string, elem string) []string {
	if slices.Contains(arr, elem) {
		log.Fatalf("version %s has already existed", elem)
	}
	return append(arr, elem)
}

// Read reads version mappings from a file and initializes the Versions struct
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

	v := &Versions{
		MetadataMap: m,
		fileName:    f.Name(),
		cVerToGoVer: map[string]*cVerMap{},
	}
	v.build()
	return v
}

// build constructs the cVerToGoVer map from the metadata
func (v *Versions) build() {
	// O(n)
	for clib := range v.MetadataMap {
		cverMap := newCverMap()

		versions := v.MetadataMap[clib]
		for _, version := range versions.VersionMappings {
			cverMap.Set(version)
		}

		v.cVerToGoVer[clib] = cverMap
	}
}

// queryClibVersion finds or creates a VersionMapping for the given C library and version
func (v *Versions) queryClibVersion(clib, clibVersion string) (versions *metadata.VersionMapping, needCreate bool) {
	versions = v.cVerToGoVer[clib].Get(clibVersion)
	// fast-path: we have a cache
	if versions != nil {
		return
	}
	needCreate = true
	// we find noting, make a blank one.
	versions = &metadata.VersionMapping{CVersion: clibVersion}
	return
}

func (v *Versions) CVersions(clib string) []string {
	return v.cVerToGoVer[clib].Versions()
}

func (v *Versions) GoVersions(clib string) []string {
	return v.cVerToGoVer[clib].GoVersions()
}

func (v *Versions) LatestGoVersionForCVersion(clib, cver string) string {
	return v.cVerToGoVer[clib].LatestGoVersionForCVersion(cver)
}

func (v *Versions) SearchBySemVer(clib, semver string) string {
	return v.cVerToGoVer[clib].SearchBySemVer(semver)
}

// LatestGoVersion returns the latest Go version associated with the given C library
func (v *Versions) LatestGoVersion(clib string) string {
	clibVer := v.cVerToGoVer[clib].LatestGoVersion()
	log.Println(clibVer)
	if clibVer != "" {
		return clibVer
	}
	allVersions := v.MetadataMap[clib]
	if allVersions == nil {
		return ""
	}
	var tmp []string
	for _, verions := range allVersions.VersionMappings {
		tmp = append(tmp, verions.GoVersions...)
	}
	if len(tmp) == 0 {
		return ""
	}
	semver.Sort(tmp)
	return tmp[len(tmp)-1]
}

// Write records a new Go version mapping for a C library version and persists to file
func (v *Versions) Write(clib, clibVersion, mappedVersion string) {
	versions, needCreate := v.queryClibVersion(clib, clibVersion)

	versions.GoVersions = appendVersion(versions.GoVersions, mappedVersion)

	if needCreate {
		if v.MetadataMap[clib] == nil {
			v.MetadataMap[clib] = &metadata.Metadata{}
		}
		v.MetadataMap[clib].VersionMappings = append(v.MetadataMap[clib].VersionMappings, versions)
	}
	// sync to disk
	b, _ := json.Marshal(&v.MetadataMap)

	os.WriteFile(v.fileName, []byte(b), 0644)
}

func (v *Versions) String() string {
	b, _ := json.Marshal(&v.MetadataMap)
	return string(b)
}
