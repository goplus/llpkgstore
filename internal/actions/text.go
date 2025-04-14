package actions

import (
	"fmt"
	"regexp"
)

const _unspecificMappedVersionRegex = `Release-as:\s%s/v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`

// regex compiles a regular expression pattern to detect "Release-as" directives in commit messages
func compileCommitVersionRegexByName(packageName string) *regexp.Regexp {
	// format: Release-as: clib/semver(with v prefix)
	// Must have one space in the end of Release-as:
	return regexp.MustCompile(fmt.Sprintf(_unspecificMappedVersionRegex, packageName))
}
