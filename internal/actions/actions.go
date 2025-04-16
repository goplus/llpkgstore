package actions

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/versions"
	"golang.org/x/mod/semver"
)

// GitHubEvent caches parsed GitHub event data from GITHUB_EVENT_PATH
var GitHubEvent = sync.OnceValue(parseGitHubEvent)

// In our previous design, each platform should generate *_{OS}_{Arch}.go file
// Feb 12th, this design revoked, still keep the code.
var currentSuffix = runtime.GOOS + "_" + runtime.GOARCH

// must panics if the error is non-nil, halting execution
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// parseGitHubEvent parses the GitHub event payload from GITHUB_EVENT_PATH into a map
func parseGitHubEvent() map[string]any {
	eventFileName := os.Getenv("GITHUB_EVENT_PATH")
	if eventFileName == "" {
		panic("cannot get GITHUB_EVENT_PATH")
	}
	event, err := os.ReadFile(eventFileName)
	if err != nil {
		panic(err)
	}
	var m map[string]any
	json.Unmarshal([]byte(event), &m)

	if len(m) == 0 {
		panic("cannot parse GITHUB_EVENT_PATH")
	}
	return m
}

// PullRequestEvent extracts pull request details from the parsed GitHub event data
func PullRequestEvent() map[string]any {
	pullRequest, ok := GitHubEvent()["pull_request"].(map[string]any)
	if !ok {
		panic("cannot parse GITHUB_EVENT_PATH pull_request")
	}
	return pullRequest
}

// IssueEvent retrieves issue-related information from the GitHub event payload
func IssueEvent() map[string]any {
	issue, ok := GitHubEvent()["issue"].(map[string]any)
	if !ok {
		panic("cannot parse GITHUB_EVENT_PATH pull_request")
	}
	return issue
}

// tagRef constructs full Git tag reference string (e.g. "refs/tags/v1.0.0")
func tagRef(tag string) string {
	return "refs/tags/" + strings.TrimSpace(tag)
}

// branchRef generates full Git branch reference string (e.g. "refs/heads/main")
func branchRef(branchName string) string {
	return "refs/heads/" + strings.TrimSpace(branchName)
}

// hasTag checks if specified Git tag exists in repository
func hasTag(tag string) bool {
	_, err := exec.Command("git", "rev-parse", tagRef(tag)).CombinedOutput()
	return err == nil
}

// shaFromTag retrieves commit SHA for given Git tag
// Panics if tag doesn't exist
func shaFromTag(tag string) string {
	ret, err := exec.Command("git", "rev-list", "-n", "1", tag).CombinedOutput()
	if err != nil {
		log.Fatalf("cannot find a tag: %s %s", tag, string(ret))
	}
	return strings.TrimSpace(string(ret))
}

// parseMappedVersion splits the mapped version string into library name and version.
// Input format: "clib/semver" where semver starts with 'v'
// Panics if input format is invalid or version isn't valid semantic version
func parseMappedVersion(version string) (clib, mappedVersion string) {
	arr := strings.Split(version, "/")
	if len(arr) != 2 {
		panic("invalid mapped version format")
	}
	clib, mappedVersion = arr[0], arr[1]

	if !semver.IsValid(mappedVersion) {
		panic("invalid mapped version format: mappedVersion is not a semver")
	}
	return
}

// isValidLLPkg checks if directory contains both llpkg.cfg and llcppg.cfg
func isValidLLPkg(files []os.DirEntry) bool {
	fileMap := make(map[string]struct{}, len(files))

	for _, file := range files {
		fileMap[filepath.Base(file.Name())] = struct{}{}
	}
	_, hasLLPkg := fileMap["llpkg.cfg"]
	_, hasLLCppg := fileMap["llcppg.cfg"]
	return hasLLCppg && hasLLPkg
}

// checkLegacyVersion validates versioning strategy for legacy package submissions
// Ensures semantic versioning compliance and proper branch maintenance strategy
func checkLegacyVersion(ver *versions.Versions, cfg config.LLPkgConfig, mappedVersion string, isLegacy bool) {
	if slices.Contains(ver.GoVersions(cfg.Upstream.Package.Name), mappedVersion) {
		panic("repeat semver")
	}
	vers := ver.CVersions(cfg.Upstream.Package.Name)
	currentVersion := versions.ToSemVer(cfg.Upstream.Package.Version)

	// skip when we're the only latest version or C version doesn't follow semver.
	if len(vers) == 0 || !semver.IsValid(currentVersion) {
		return
	}

	sort.Sort(versions.ByVersionDescending(vers))

	latestVersion := vers[0]

	isLatest := semver.Compare(currentVersion, latestVersion) > 0
	// fast-path: we're the latest version
	if isLatest {
		// case1: we're the latest version, but mapped version is not latest, invalid.
		// example: all version: 1.8.1 => v1.2.0 1.7.1 => v1.1.0 current: 1.9.1 => v1.0.0
		if semver.Compare(ver.LatestGoVersion(cfg.Upstream.Package.Name), mappedVersion) > 0 {
			panic("mapped version should not less than the legacy one.")
		}
		return
	} else if !isLegacy {
		// case2: if we're legacy version, the pr is submited to main, that's invalid.
		// in the most common case, it should be conflict.
		// however, consider about the extraordinary case.
		panic("legacy version MUST not submit to main branch")
	}

	// find the closest verion which is smaller than us.
	i := sort.Search(len(vers), func(i int) bool {
		return semver.Compare(vers[i], currentVersion) < 0
	})

	hasClosestSemver := i < len(vers) &&
		semver.Compare(vers[i], currentVersion) < 0
	// case3: we're the smallest version
	// example: latest: 1.6.1 maintain: 1.5.1, that's valid
	if !hasClosestSemver {
		return
	}

	// case4: the major and minor version of the previous version is same,
	// which means we're not the latest patch version, invalid.
	// example: all version: 1.6.1 1.5.3 1.5.1 current: 1.5.2, so the previous one is 1.5.3, that's invalid
	previousVersion := vers[i-1]

	if semver.MajorMinor(previousVersion) == semver.MajorMinor(currentVersion) &&
		semver.Compare(previousVersion, currentVersion) > 0 {
		panic(`cannot submit a historical legacy version.
	for more details: https://github.com/goplus/llpkgstore/blob/main/docs/llpkgstore.md#branch-maintenance-strategy`)
	}

	// case5: we're the latest patch version for current major and minor, check the mapped version
	// our mapped version should be larger than the closest one.
	// example: current submit: 1.5.2 => v1.1.1, closest minor: 1.4.1 => v1.1.0, valid.
	originalVersion := ver.SearchBySemVer(cfg.Upstream.Package.Name, vers[i])
	if originalVersion == "" {
		panic("cannot find original C version from semver, this should not happen.")
	}
	closestMappedVersion := ver.LatestGoVersionForCVersion(cfg.Upstream.Package.Name, originalVersion)
	if closestMappedVersion == "" {
		panic("cannot find latest Go version from C version, this should not happen.")
	}

	if semver.Compare(closestMappedVersion, mappedVersion) > 0 {
		panic("mapped version should not less than the legacy one.")
	}
}
