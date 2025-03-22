package versions

import "golang.org/x/mod/semver"

// Package versions provides utilities for working with semantic versioning.

// ByVersionDescending implements [sort.Interface] for sorting semantic version strings in descending order.
type ByVersionDescending []string

func (vs ByVersionDescending) Len() int      { return len(vs) }
func (vs ByVersionDescending) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }

func (vs ByVersionDescending) Less(i, j int) bool {
	cmp := semver.Compare(vs[i], vs[j])
	if cmp != 0 {
		return cmp > 0
	}
	return vs[i] > vs[j]
}

// fillVersionPrefix ensures version strings start with 'v' prefix
func fillVersionPrefix(version string) string {
	if version == "" {
		panic("invalid empty version")
	}
	if version[0] == 'v' {
		return version
	}
	return "v" + version
}

// ToSemVer converts a version string to canonical semantic version format
func ToSemVer(version string) string {
	semver := semver.Canonical(fillVersionPrefix(version))
	if semver == "" {
		return version
	}
	return semver
}

// IsSemver checks if all provided version strings are valid semantic versions
func IsSemver(cversions []string) bool {
	for _, cversion := range cversions {
		if !semver.IsValid(cversion) {
			return false
		}
	}
	return true
}
