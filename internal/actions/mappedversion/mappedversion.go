package mappedversion

import (
	"errors"
	"strings"

	"golang.org/x/mod/semver"
)

var (
	ErrVersionFormat = errors.New("invalid mapped version format: mappedVersion is not a semver")
)

type MappedVersion string

func From(version string) MappedVersion {
	return MappedVersion(version)
}

// Parse splits the mapped version string into library name and version.
// Input format: "clib/semver" where semver starts with 'v'
// Panics if input format is invalid or version isn't valid semantic version
func (m MappedVersion) Parse() (clib, mappedVersion string, err error) {
	arr := strings.Split(string(m), "/")
	if len(arr) != 2 {
		panic("invalid mapped version format")
	}
	clib, mappedVersion = arr[0], arr[1]

	if !semver.IsValid(mappedVersion) {
		err = ErrVersionFormat
	}
	return
}

func (m MappedVersion) MustParse() (clib, mappedVersion string) {
	clib, mappedVersion, err := m.Parse()
	if err != nil {
		panic(err)
	}
	return
}
