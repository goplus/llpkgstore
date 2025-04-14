package mappedversion

import (
	"errors"
	"strings"

	"golang.org/x/mod/semver"
)

// ErrVersionFormat invalid version format error
var ErrVersionFormat = errors.New("invalid mapped version format")

// ErrMappedVersionFormat invalid semantic version error
var ErrMappedVersionFormat = errors.New("invalid semantic version format")

type MappedVersion string

// From creates a MappedVersion from string
func From(version string) MappedVersion {
	return MappedVersion(version)
}

// Parse splits the mapped version string into library name and version.
// Input format: "clib/semver" where semver starts with 'v'
// Panics if input format is invalid or version isn't valid semantic version
func (m MappedVersion) Parse() (clib, version string, err error) {
	parts := strings.Split(string(m), "/")
	if len(parts) != 2 {
		return "", "", ErrVersionFormat
	}
	clib, version = parts[0], parts[1]
	if !semver.IsValid(version) {
		return "", "", ErrMappedVersionFormat
	}
	return
}

// MustParse parses version or panics
func (m MappedVersion) MustParse() (string, string) {
	clib, ver, err := m.Parse()
	if err != nil {
		panic(err)
	}
	return clib, ver
}

// String returns the version string representation
func (m MappedVersion) String() string {
	return string(m)
}
