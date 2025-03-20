package versions

import (
	"github.com/goplus/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

type CVerMap map[string]*metadata.VersionMapping

func (c CVerMap) GoVersions() (tmp []string) {
	if len(c) == 0 {
		return nil
	}
	for _, verions := range c {
		tmp = append(tmp, verions.GoVersions...)
	}
	return
}

func (c CVerMap) LatestGoVersion() string {
	if len(c) == 0 {
		return ""
	}
	tmp := c.GoVersions()
	// check again to avoid unexpected behavior
	if len(tmp) == 0 {
		return ""
	}
	semver.Sort(tmp)
	return tmp[len(tmp)-1]
}

func (c CVerMap) Get(cver string) *metadata.VersionMapping {
	// it's possible that we're uninitiated
	if len(c) == 0 {
		return nil
	}
	return c[cver]
}

func (c CVerMap) LatestGoVersionForCVersion(cver string) string {
	mappingTable := c.Get(cver)
	if mappingTable == nil || len(mappingTable.GoVersions) == 0 {
		return ""
	}
	goVersions := make([]string, len(mappingTable.GoVersions))
	copy(goVersions, mappingTable.GoVersions)
	semver.Sort(goVersions)

	return goVersions[len(goVersions)-1]
}

func (c CVerMap) SearchBySemVer(semver string) (originalVer string) {
	for cver := range c {
		if ToSemVer(cver) == semver {
			originalVer = cver
			break
		}
	}
	return
}

func (c CVerMap) Versions() (ret []string) {
	for version := range c {
		ret = append(ret, ToSemVer(version))
	}
	return
}
