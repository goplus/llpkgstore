package versions

import (
	"encoding/json"
	"io"
	"os"
	"slices"

	"github.com/goplus/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

// Versions is a mapping table wrapper for Github Action only.
// It's recommend to use another implement in llgo for common usage.
type Versions struct {
	metadata.MetadataMap

	fileName    string
	cVerToGoVer map[string]CVerMap
}

// appendUnique adds a unique element to a slice.
func appendUnique(arr []string, elem string) []string {
	arr = append(arr, elem)
	slices.Sort(arr)
	return slices.Compact(arr)
}

// ReadVersion reads version mappings from a file and initializes the Versions struct
func ReadVersion(fileName string) *Versions {
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
		cVerToGoVer: map[string]CVerMap{},
	}
	v.build()
	return v
}

// build constructs the cVerToGoVer map from the metadata
func (v *Versions) build() {
	// O(n)
	for clib := range v.MetadataMap {
		cverMap := CVerMap{}

		versions := v.MetadataMap[clib]
		for _, version := range versions.VersionMappings {
			cverMap[version.CVersion] = version
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
	// slow-path: parse it
	allVersions := v.MetadataMap[clib]
	if allVersions != nil {
		for _, mapping := range allVersions.VersionMappings {
			if mapping.CVersion == clibVersion {
				versions = mapping
				return
			}
		}
	}
	needCreate = true
	// we find noting, make a blank one.
	versions = &metadata.VersionMapping{CVersion: clibVersion}
	return
}

// LatestGoVersion returns the latest Go version associated with the given C library
func (v *Versions) LatestGoVersion(clib string) string {
	clibVer := v.cVerToGoVer[clib].LatestGoVersion()
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

	versions.GoVersions = appendUnique(versions.GoVersions, mappedVersion)

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
