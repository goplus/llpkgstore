package versions

import (
	"github.com/MeteorsLiu/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

type CVerMap map[string]*metadata.VersionMapping

func (c CVerMap) LatestGoVersion() string {
	if len(c) == 0 {
		return ""
	}
	var tmp []string
	for _, verions := range c {
		tmp = append(tmp, verions.GoVersions...)
	}
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
