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

func appendUnique(arr []string, elem string) []string {
	arr = append(arr, elem)
	slices.Sort(arr)
	return slices.Compact(arr)
}

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

func (v *Versions) queryClibVersion(clib, clibVersion string) (versions *metadata.VersionMapping, needCreate bool) {
	versions = v.cVerToGoVer[clib].Get(clibVersion)
	// fast-path: we have a cache
	if versions != nil {
		return
	}
	// slow-path: parse it
	allVersions := v.MetadataMap[clib]
	for _, mapping := range allVersions.VersionMappings {
		if mapping.CVersion == clibVersion {
			versions = mapping
			return
		}
	}
	needCreate = true
	// we find noting, make a blank one.
	versions = &metadata.VersionMapping{CVersion: clibVersion}
	return
}

func (v *Versions) LatestGoVersion(clib string) string {
	clibVer := v.cVerToGoVer[clib].LatestGoVersion()
	if clibVer != "" {
		return clibVer
	}
	allVersions := v.MetadataMap[clib]
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

func (v *Versions) Write(clib, clibVersion, mappedVersion string) {
	versions, needCreate := v.queryClibVersion(clib, clibVersion)

	versions.GoVersions = appendUnique(versions.GoVersions, mappedVersion)

	if needCreate {
		current := v.MetadataMap[clib]
		current.VersionMappings = append(current.VersionMappings, versions)
		v.MetadataMap[clib] = current
	}
	// sync to disk
	b, _ := json.Marshal(&v.MetadataMap)

	os.WriteFile(v.fileName, []byte(b), 0644)
}
