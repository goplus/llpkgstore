package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/goplus/llpkgstore/actions/versions"
	"github.com/goplus/llpkgstore/config"
	"golang.org/x/mod/semver"
)

var GithubEvent = sync.OnceValue(parseGithubEvent)

// In our previous design, each platform should generate *_{OS}_{Arch}.go file
// Feb 12th, this design revoked, still keep the code.
// var currentSuffix = runtime.GOOS + "_" + runtime.GOARCH

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// envToString converts a env map to string
func envToString(envm map[string]string) string {
	var env []string

	for name, value := range envm {
		env = append(env, fmt.Sprintf("%s=%s", name, value))
	}
	return strings.Join(env, "\n")
}

func parseGithubEvent() map[string]any {
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

func PullRequestEvent() map[string]any {
	pullRequest, ok := GithubEvent()["pull_request"].(map[string]any)
	if !ok {
		panic("cannot parse GITHUB_EVENT_PATH pull_request")
	}
	return pullRequest
}

func IssueEvent() map[string]any {
	issue, ok := GithubEvent()["issue"].(map[string]any)
	if !ok {
		panic("cannot parse GITHUB_EVENT_PATH pull_request")
	}
	return issue
}

// tagRef returns full Git ref for a tag (e.g. "refs/tags/v1.0.0")
func tagRef(tag string) string {
	return "refs/tags/" + strings.TrimSpace(tag)
}

// branchRef returns full Git ref for a branch (e.g. "refs/heads/main")
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
	ret, _ := exec.Command("git", "tag").CombinedOutput()
	log.Println(string(ret))
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

// isValidLlpkg checks if directory contains both llpkg.cfg and llcppg.cfg
func isValidLlpkg(files []string) bool {
	fileMap := make(map[string]struct{}, len(files))

	for _, file := range files {
		fileMap[file] = struct{}{}
	}
	_, hasLlpkg := fileMap["llpkg.cfg"]
	_, hasLlcppg := fileMap["llcppg.cfg"]
	return hasLlcppg && hasLlpkg
}

// isLegacyVersion reports current PR stands for legacy version
func isLegacyVersion() (branchName string, legacy bool) {
	pullRequest := PullRequestEvent()

	// unnecessary to check type, because currentPRCommit has been checked.
	base := pullRequest["base"].(map[string]any)
	refName := base["ref"].(string)

	legacy = strings.HasPrefix(refName, BranchPrefix)
	branchName = refName
	return
}

func checkLegacyVersion(ver *versions.Versions, cfg config.LLPkgConfig, mappedVersion string) {
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
	// case1: if this is a legacy version, it MUST be not in main
	_, isLegacy := isLegacyVersion()

	if semver.Compare(currentVersion, latestVersion) < 0 && !isLegacy {
		panic("legacy version MUST not submit to main branch")
	}

	// find the latest minor verion
	// cloest semver: same major and minor semver
	i := sort.Search(len(vers), func(i int) bool {
		return semver.MajorMinor(vers[i]) == semver.MajorMinor(currentVersion)
	})

	hasClosestSemver := i < len(vers) &&
		semver.MajorMinor(vers[i]) == semver.MajorMinor(currentVersion)
	// case2: the cloest semver not found
	// example: latest: 1.6.1 maintain: 1.5.1, that's valid
	if !hasClosestSemver {
		// case 4: if we're the latest version, check mapped version.(TestCase 4)
		// example: latest: 1.8.1 => v0.2.0, 1.9.1 => v0.1.0, invalid!
		if semver.Compare(currentVersion, latestVersion) > 0 &&
			semver.Compare(mappedVersion, ver.LatestGoVersion(cfg.Upstream.Package.Name)) <= 0 {
			panic("mapped version cannot less than the legacy one")
		}
		// case 5: we're not the latest version and not the legacy version
		// example: legacy: 1.6.1 => v0.3.0 1.4.1 => v0.1.0, current: 1.5.1 => v0.2.0
		// we check nothing here.
		return
	}

	closestSemver := vers[i]

	originalVersion := ver.SearchBySemVer(cfg.Upstream.Package.Name, vers[i])
	if originalVersion == "" {
		panic("cannot find original C version from semver, this should not happen.")
	}
	closestMappedVersion := ver.LatestGoVersionForCVersion(cfg.Upstream.Package.Name, originalVersion)
	if closestMappedVersion == "" {
		panic("cannot find latest Go version from C version, this should not happen.")
	}

	// case3: we're the latest patch version, that's valid
	// example: current submit: 1.5.2 => v1.1.1, closest minor: 1.5.1 => v1.1.0, valid.
	if semver.Compare(currentVersion, closestSemver) > 0 &&
		semver.Compare(mappedVersion, closestMappedVersion) > 0 {
		return
	}

	panic(`cannot submit a historical legacy version.
	for more details: https://github.com/goplus/llpkgstore/blob/main/docs/llpkgstore.md#branch-maintenance-strategy`)
}

// Setenv sets the value of the Github Action environment variable named by the key.
func Setenv(envm map[string]string) {
	env, err := os.OpenFile(os.Getenv("GITHUB_ENV"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	// should never happen,
	// it means current runtime is not Github actions if there's any errors
	must(err)

	env.WriteString(envToString(envm))

	// make sure we write it to the GITHUB_ENV
	env.Close()
}

// SetOutput sets the value of the Github Action workflow output named by the key.
func SetOutput(envm map[string]string) {
	env, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	must(err)

	env.WriteString(envToString(envm))

	env.Close()
}

// Changes returns the changed files in current PR,
// which depends on ALL_CHANGED_FILES generated by tj-actions/changed-files action,
// if there's no content in ALL_CHANGED_FILES, it panic.
func Changes() []string {
	changes := os.Getenv("ALL_CHANGED_FILES")
	if changes == "" {
		panic("cannot find changes file!")
	}
	return strings.Fields(changes)
}

// Repository returns owner and repository name for the current repository
//
// Example: goplus/llpkg, owner: goplus, repo: llpkg
func Repository() (owner, repo string) {
	thisRepo := os.Getenv("GITHUB_REPOSITORY")
	if thisRepo == "" {
		panic("no github repo")
	}
	current := strings.Split(thisRepo, "/")
	return current[0], current[1]
}

// Token returns Github Token for current runner
func Token() string {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		panic("no GH_TOKEN")
	}
	return token
}
