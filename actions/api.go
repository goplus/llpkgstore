package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/goplus/llpkgstore/actions/versions"
	"github.com/goplus/llpkgstore/config"
	"golang.org/x/mod/semver"
)

const (
	LabelPrefix         = "branch:"
	BranchPrefix        = "release-branch."
	MappedVersionPrefix = "Release-as: "
	regexString         = `Release-as:\s%s/v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

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

func tagRef(tag string) string {
	return "refs/tags/" + strings.TrimSpace(tag)
}

func branchRef(branchName string) string {
	return "refs/heads/" + strings.TrimSpace(branchName)
}

func hasTag(tag string) bool {
	_, err := exec.Command("git", "rev-parse", tagRef(tag)).CombinedOutput()
	return err == nil
}

func shaFromTag(tag string) string {
	ret, err := exec.Command("git", "rev-list", "-n", "1", tag).CombinedOutput()
	if err != nil {
		log.Fatalf("cannot find a tag: %s", tag)
	}
	return strings.TrimSpace(string(ret))
}

func regex(packageName string) *regexp.Regexp {
	// format: Release-as: clib/semver(with v prefix)
	// Must have one space in the end of Release-as:
	return regexp.MustCompile(fmt.Sprintf(regexString, packageName))
}

func isValidLlpkg(files []string) bool {
	fileMap := make(map[string]struct{}, len(files))

	for _, file := range files {
		fileMap[file] = struct{}{}
	}
	_, hasLlpkg := fileMap["llpkg.cfg"]
	_, hasLlcppg := fileMap["llcppg.cfg"]
	return hasLlcppg && hasLlpkg
}

type DefaultClient struct {
	repo   string
	owner  string
	client *github.Client
}

func NewDefaultClient() *DefaultClient {
	dc := &DefaultClient{
		client: github.NewClient(nil).WithAuthToken(Token()),
	}
	dc.owner, dc.repo = Repository()
	return dc
}

func (d *DefaultClient) hasBranch(branchName string) bool {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branch, resp, err := d.client.Repositories.GetBranch(
		ctx, d.owner, d.repo, branchName, 0,
	)
	if err != nil {
		panic(err)
	}
	return branch != nil &&
		resp.StatusCode == http.StatusOK
}

func (d *DefaultClient) isAssociatedWithPullRequest(sha string) bool {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	pulls, _, err := d.client.PullRequests.ListPullRequestsWithCommit(
		ctx, d.owner, d.repo, sha, &github.ListOptions{},
	)
	if err != nil {
		panic(err)
	}
	// don't use GetMerge, because GetMerge may be a mistake.
	// sometime, when a pull request is merged, GetMerge still returns false.
	// so checking pull request state is more accurate.
	return len(pulls) > 0 &&
		pulls[0].GetState() == "closed"
}

// currentPRCommit returns all the commits for the current PR.
func (d *DefaultClient) currentPRCommit() []*github.RepositoryCommit {
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
	pullRequest, ok := m["pull_request"].(map[string]any)
	if !ok {
		panic("cannot parse GITHUB_EVENT_PATH pull_request")
	}
	prNumber := int(pullRequest["number"].(float64))

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	commits, _, err := d.client.PullRequests.ListCommits(
		ctx, d.owner, d.repo, prNumber,
		&github.ListOptions{},
	)
	if err != nil {
		panic(err)
	}
	return commits
}

func (d *DefaultClient) checkMappedVersion(packageName string) {
	matchMappedVersion := regex(packageName)

	found := false
	for _, commit := range d.currentPRCommit() {
		message := commit.GetCommit().GetMessage()
		if matchMappedVersion.Match([]byte(message)) {
			found = true
			break
		}
	}

	if !found {
		panic("no MappedVersion found in the PR")
	}
}

func (d *DefaultClient) commitMessage(sha string) *github.RepositoryCommit {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	commit, _, err := d.client.Repositories.GetCommit(ctx, d.owner, d.repo, sha, &github.ListOptions{})
	if err != nil {
		panic(err)
	}
	return commit
}

func (d *DefaultClient) mappedVersion() string {
	// get message
	message := d.commitMessage(os.Getenv("GITHUB_SHA")).GetCommit().GetMessage()

	// parse the mapped version
	mappedVersion := regex(".*").FindString(message)

	// mapped version not found, a normal commit?
	if mappedVersion == "" {
		return ""
	}
	version := strings.TrimPrefix(mappedVersion, MappedVersionPrefix)
	if version == mappedVersion {
		panic("invalid format")
	}
	return version
}

func (d *DefaultClient) createTag(tag, sha string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	// tag the commit
	tagRefName := tagRef(tag)
	_, _, err := d.client.Git.CreateRef(ctx, d.owner, d.repo, &github.Reference{
		Ref: &tagRefName,
		Object: &github.GitObject{
			SHA: &sha,
		},
	})

	return err
}

func (d *DefaultClient) createBranch(branchName, sha string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branchRefName := branchRef(branchName)
	_, _, err := d.client.Git.CreateRef(ctx, d.owner, d.repo, &github.Reference{
		Ref: &branchRefName,
		Object: &github.GitObject{
			SHA: &sha,
		},
	})

	return err
}

func (d *DefaultClient) CheckPR() []string {
	// build a file path map
	pathMap := map[string][]string{}
	for _, path := range Changes() {
		dir := filepath.Dir(path)
		pathMap[dir] = append(pathMap[dir], filepath.Base(path))
	}
	var packages []string

	for path := range pathMap {
		files := pathMap[path]

		if !isValidLlpkg(files) {
			delete(pathMap, path)
			continue
		}
		// 3. Check directory name
		llpkgFile := filepath.Join(path, "llpkg.cfg")
		cfg, err := config.ParseLLPkgConfig(llpkgFile)
		if err != nil {
			panic(err)
		}
		// in our design, directory name should equal to the package name,
		// which means it's not required to be equal.
		//
		// However, at the current stage, if this is not equal, conan may panic,
		// to aovid unexpected behavior, we assert it's equal temporarily.
		// this logic may be changed in the future.
		packageName := strings.TrimSpace(cfg.Upstream.Package.Name)
		if packageName != path {
			panic("directory name is not equal to package name in llpkg.cfg")
		}
		packages = append(packages, packageName)
	}

	// 1. Check there's only one directory in PR
	if len(pathMap) > 1 {
		panic("too many to-be-converted directory")
	}

	// 2. Check config files(llpkg.cfg and llcppg.cfg)
	if len(pathMap) == 0 {
		panic("no valid config files, llpkg.cfg and llcppg.cfg must exist")
	}

	// 4. Check MappedVersion
	//
	// it should be one package name at the current stage,
	// however, in the future, it may allow multiple packages.
	for _, packageName := range packages {
		d.checkMappedVersion(packageName)
	}

	return slices.Collect(maps.Keys(pathMap))
}

func (d *DefaultClient) Release() {
	// https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#push
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		panic("no GITHUB_SHA found")
	}
	// check it's associated with a pr
	if !d.isAssociatedWithPullRequest(sha) {
		// not a merge commit, skip it.
		return
	}

	version := d.mappedVersion()
	// skip it when no mapped version is found
	if version == "" {
		return
	}

	if hasTag(version) {
		// tag existed already, skip it.
		return
	}

	if err := d.createTag(version, sha); err != nil {
		panic(err)
	}

	clib, mappedVersion := parseMappedVersion(version)

	// the pr has merged, so we can read it.
	config, err := config.ParseLLPkgConfig(filepath.Join(clib, "llpkg.cfg"))
	if err != nil {
		panic(err)
	}

	// write it to llpkgstore.json
	ver := versions.ReadVersion("llpkgstore.json")
	ver.Write(clib, config.Upstream.Package.Version, mappedVersion)

	// move to website in Github Action...
}

func (d *DefaultClient) CreateBranchFromLabel(labelName string) {
	// design: branch:release-branch.{CLibraryName}/{MappedVersion}
	branchName := strings.TrimPrefix(labelName, LabelPrefix)
	if branchName == labelName {
		panic("invalid label name format")
	}

	// fast-path: branch exists, can skip.
	if d.hasBranch(branchName) {
		return
	}
	version := strings.TrimPrefix(labelName, BranchPrefix)
	if version == labelName {
		panic("invalid label name format")
	}
	clib, mappedVersion := parseMappedVersion(version)
	// slow-path: check the condition if we can create a branch
	//
	// create a branch only when this version is legacy.
	// according to branch maintenance strategy

	// step 1: get latest version of the clib
	ver := versions.ReadVersion("llpkgstore.json")
	latestVersion := ver.LatestGoVersion(clib)

	// unnecessary to create a branch if mappedVersion >= latestVersion
	if semver.Compare(mappedVersion, latestVersion) >= 0 {
		return
	}

	if err := d.createBranch(branchName, shaFromTag(version)); err != nil {
		panic(err)
	}
}
