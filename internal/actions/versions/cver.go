package versions

import (
	"github.com/goplus/llpkgstore/metadata"
	"golang.org/x/mod/semver"
)

// cVerMap represents a map of C versions to their corresponding version mappings
type cVerMap struct {
	m map[string]*metadata.VersionMapping

	cVerCache       []string // cached list of C versions
	goVersionsCache []string // cached list of Go versions
}

// newCverMap creates a new empty cVerMap instance
func newCverMap() *cVerMap {
	return &cVerMap{m: map[string]*metadata.VersionMapping{}}
}

// GoVersions returns all Go versions associated with all C versions
func (c *cVerMap) GoVersions() (tmp []string) {
	if c == nil || len(c.m) == 0 {
		return nil
	}
	if c.goVersionsCache != nil {
		tmp = c.goVersionsCache
		return
	}
	for _, versions := range c.m { // Fixed typo: verions -> versions
		tmp = append(tmp, versions.GoVersions...)
	}
	c.goVersionsCache = tmp // Cache computed value
	return
}

// Versions returns all C versions stored in the map
func (c *cVerMap) Versions() []string {
	if c == nil || len(c.m) == 0 {
		return nil
	}
	if c.cVerCache != nil {
		return c.cVerCache
	}
	var tmp []string
	for version := range c.m {
		tmp = append(tmp, version)
	}
	c.cVerCache = tmp // Cache computed value
	return tmp
}

// LatestGoVersion returns the highest semantic version of all Go versions
func (c *cVerMap) LatestGoVersion() string {
	if c == nil || len(c.m) == 0 {
		return ""
	}
	goVersions := c.GoVersions()
	if len(goVersions) == 0 {
		return ""
	}
	semver.Sort(goVersions)
	return goVersions[len(goVersions)-1]
}

// Get retrieves the VersionMapping for a specific C version
func (c *cVerMap) Get(cver string) *metadata.VersionMapping {
	if c == nil || len(c.m) == 0 {
		return nil
	}
	return c.m[cver]
}

// Set adds or updates a VersionMapping entry
func (c *cVerMap) Set(versions *metadata.VersionMapping) {
	c.m[versions.CVersion] = versions
	c.cVerCache = append(c.cVerCache, ToSemVer(versions.CVersion))        // Update C version cache
	c.goVersionsCache = append(c.goVersionsCache, versions.GoVersions...) // Update Go versions cache
}

// LatestGoVersionForCVersion returns the latest Go version for a specific C version
func (c *cVerMap) LatestGoVersionForCVersion(cver string) string {
	mappingTable := c.Get(cver)
	if mappingTable == nil || len(mappingTable.GoVersions) == 0 {
		return ""
	}
	goVersions := make([]string, len(mappingTable.GoVersions))
	copy(goVersions, mappingTable.GoVersions) // Fixed typo: GoVersion -> GoVersions
	semver.Sort(goVersions)
	return goVersions[len(goVersions)-1]
}

// SearchBySemVer finds the original C version from its semantic version string
func (c *cVerMap) SearchBySemVer(semver string) (originalVer string) {
	for cver := range c.m {
		if ToSemVer(cver) == semver {
			originalVer = cver
			break
		}
	}
	return
}
