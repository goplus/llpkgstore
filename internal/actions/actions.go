package actions

import (
	"encoding/json"
	"fmt"
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
	"github.com/goplus/llpkgstore/internal/actions/env"
	"github.com/goplus/llpkgstore/internal/actions/versions"
	"github.com/goplus/llpkgstore/internal/file"
	"github.com/goplus/llpkgstore/internal/pc"
	"github.com/goplus/llpkgstore/upstream"
	"golang.org/x/mod/semver"
)

// GitHubEvent caches parsed GitHub event data from GITHUB_EVENT_PATH
var GitHubEvent = sync.OnceValues(parseGitHubEvent)

// In our previous design, each platform should generate *_{OS}_{Arch}.go file
// Feb 12th, this design revoked, still keep the code.
var currentSuffix = runtime.GOOS + "_" + runtime.GOARCH

// parseGitHubEvent parses the GitHub event payload from GITHUB_EVENT_PATH into a map
func parseGitHubEvent() (map[string]any, error) {
	eventFile, err := env.EventFile()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(eventFile, &m)

	if err != nil {
		return nil, err
	}
	return m, nil
}

// PullRequestEvent extracts pull request details from the parsed GitHub event data
func PullRequestEvent() (map[string]any, error) {
	event, err := GitHubEvent()
	if err != nil {
		return nil, err
	}
	pullRequest, ok := event["pull_request"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("env: cannot parse GITHUB_EVENT_PATH pull_request")
	}
	return pullRequest, nil
}

// IssueEvent retrieves issue-related information from the GitHub event payload
func IssueEvent() (map[string]any, error) {
	event, err := GitHubEvent()
	if err != nil {
		return nil, err
	}
	issue, ok := event["issue"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("env: cannot parse GITHUB_EVENT_PATH pull_request")
	}
	return issue, nil
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
func parseMappedVersion(version string) (clib, mappedVersion string, err error) {
	arr := strings.Split(version, "/")
	if len(arr) != 2 {
		err = fmt.Errorf("actions: invalid mapped version format")
		return
	}
	clib, mappedVersion = arr[0], arr[1]

	if !semver.IsValid(mappedVersion) {
		err = fmt.Errorf("actions: invalid mapped version format: mappedVersion is not a semver")
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
func checkLegacyVersion(ver *versions.Versions, cfg config.LLPkgConfig, mappedVersion string, isLegacy bool) error {
	if slices.Contains(ver.GoVersions(cfg.Upstream.Package.Name), mappedVersion) {
		return fmt.Errorf("actions: repeat semver %s", mappedVersion)
	}
	vers := ver.CVersions(cfg.Upstream.Package.Name)
	currentVersion := versions.ToSemVer(cfg.Upstream.Package.Version)

	// skip when we're the only latest version or C version doesn't follow semver.
	if len(vers) == 0 || !semver.IsValid(currentVersion) {
		return nil
	}

	sort.Sort(versions.ByVersionDescending(vers))

	latestVersion := vers[0]

	isLatest := semver.Compare(currentVersion, latestVersion) >= 0
	// fast-path: we're the latest version
	if isLatest {
		// case1: we're the latest version, but mapped version is not latest, invalid.
		// example: all version: 1.8.1 => v1.2.0 1.7.1 => v1.1.0 current: 1.9.1 => v1.0.0
		if semver.Compare(ver.LatestGoVersion(cfg.Upstream.Package.Name), mappedVersion) > 0 {
			return fmt.Errorf("actions: mapped version should not less than the legacy one")
		}
		return nil
	} else if !isLegacy {
		// case2: if we're legacy version, the pr is submited to main, that's invalid.
		// in the most common case, it should be conflict.
		// however, consider about the extraordinary case.
		return fmt.Errorf("actions: legacy version MUST not submit to main branch")
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
		return nil
	}

	// case4: the major and minor version of the previous version is same,
	// which means we're not the latest patch version, invalid.
	// example: all version: 1.6.1 1.5.3 1.5.1 current: 1.5.2, so the previous one is 1.5.3, that's invalid
	previousVersion := vers[i-1]

	if semver.MajorMinor(previousVersion) == semver.MajorMinor(currentVersion) &&
		semver.Compare(previousVersion, currentVersion) > 0 {
		return fmt.Errorf(`actions: cannot submit a historical legacy version.
	for more details: https://github.com/goplus/llpkgstore/blob/main/docs/llpkgstore.md#branch-maintenance-strategy`)
	}

	// case5: we're the latest patch version for current major and minor, check the mapped version
	// our mapped version should be larger than the closest one.
	// example: current submit: 1.5.2 => v1.1.1, closest minor: 1.4.1 => v1.1.0, valid.
	originalVersion := ver.SearchBySemVer(cfg.Upstream.Package.Name, vers[i])
	if originalVersion == "" {
		return fmt.Errorf("actions: cannot find original C version from semver")
	}
	closestMappedVersion := ver.LatestGoVersionForCVersion(cfg.Upstream.Package.Name, originalVersion)
	if closestMappedVersion == "" {
		return fmt.Errorf("cannot find latest Go version from C version")
	}

	if semver.Compare(closestMappedVersion, mappedVersion) > 0 {
		return fmt.Errorf("mapped version should not less than the legacy one")
	}
	return nil
}

func BuildBinaryZip(uc *upstream.Upstream) (zipFileName, zipFilePath string, err error) {
	tempDir, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		err = wrapActionError(err)
		return
	}

	deps, err := uc.Installer.Install(uc.Pkg, tempDir)
	if err != nil {
		return
	}

	pkgConfigDir := filepath.Join(tempDir, "lib", "pkgconfig")
	// clear exist .pc
	err = os.RemoveAll(pkgConfigDir)
	if err != nil {
		err = wrapActionError(err)
		return
	}

	err = os.Mkdir(pkgConfigDir, 0777)
	if err != nil {
		err = wrapActionError(err)
		return
	}

	for _, pcName := range deps {
		pcFile := filepath.Join(tempDir, pcName+".pc")
		// generate pc template to lib/pkgconfig
		err = pc.GenerateTemplateFromPC(pcFile, pkgConfigDir, deps)
		if err != nil {
			err = wrapActionError(err)
			return
		}
	}

	// okay, safe to remove old pc
	file.RemovePattern(filepath.Join(tempDir, "*.pc"))
	file.RemovePattern(filepath.Join(tempDir, "*.sh"))

	zipFileName = binaryZip(uc.Pkg.Name)
	zipFilePath, err = filepath.Abs(zipFileName)
	if err != nil {
		err = wrapActionError(err)
		return
	}

	err = file.Zip(tempDir, zipFilePath)
	if err != nil {
		err = wrapActionError(err)
	}

	return
}
