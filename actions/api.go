// Package actions contains GitHub Actions helper functions for version management and repository operations.
package actions

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/goplus/llpkgstore/actions/versions"
	"github.com/goplus/llpkgstore/config"
)

const (
	LabelPrefix         = "branch:"
	BranchPrefix        = "release-branch."
	MappedVersionPrefix = "Release-as: "
	regexString         = `Release-as:\s%s/v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
)

// regex creates compiled regular expression for mapped version detection in commit messages
func regex(packageName string) *regexp.Regexp {
	// format: Release-as: clib/semver(with v prefix)
	// Must have one space in the end of Release-as:
	return regexp.MustCompile(fmt.Sprintf(regexString, packageName))
}

// DefaultClient provides GitHub API operations for Actions workflows
type DefaultClient struct {
	repo   string
	owner  string
	client *github.Client
}

// NewDefaultClient creates configured GitHub API client with auth token
func NewDefaultClient() *DefaultClient {
	dc := &DefaultClient{
		client: github.NewClient(nil).WithAuthToken(Token()),
	}
	dc.owner, dc.repo = Repository()
	return dc
}

// hasBranch checks if specified branch exists in repository
func (d *DefaultClient) hasBranch(branchName string) bool {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branch, resp, err := d.client.Repositories.GetBranch(
		ctx, d.owner, d.repo, branchName, 0,
	)

	return err == nil && branch != nil &&
		resp.StatusCode == http.StatusOK
}

// isAssociatedWithPullRequest checks if commit SHA is part of a closed PR
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

// currentPRCommit retrieves all commits associated with current PR
func (d *DefaultClient) currentPRCommit() []*github.RepositoryCommit {
	pullRequest := PullRequestEvent()
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

func (d *DefaultClient) allCommits() []*github.RepositoryCommit {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	commits, _, err := d.client.Repositories.ListCommits(
		ctx, d.owner, d.repo,
		&github.CommitsListOptions{},
	)
	if err != nil {
		panic(err)
	}
	return commits
}

func (d *DefaultClient) removeLabel(labelName string) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	_, err := d.client.Issues.DeleteLabel(
		ctx, d.owner, d.repo, labelName,
	)
	if err != nil {
		panic(err)
	}
	return
}

// checkMappedVersion ensures PR commit messages contain valid mapped version declaration
func (d *DefaultClient) checkMappedVersion(packageName string) (mappedVersion string) {
	matchMappedVersion := regex(packageName)

	for _, commit := range d.currentPRCommit() {
		message := commit.GetCommit().GetMessage()
		if mappedVersion = matchMappedVersion.FindString(message); mappedVersion != "" {
			// remove space, of course
			mappedVersion = strings.TrimSpace(mappedVersion)
			break
		}
	}

	if mappedVersion == "" {
		panic("no MappedVersion found in the PR")
	}
	return
}

// commitMessage retrieves commit details by SHA
func (d *DefaultClient) commitMessage(sha string) *github.RepositoryCommit {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	commit, _, err := d.client.Repositories.GetCommit(ctx, d.owner, d.repo, sha, &github.ListOptions{})
	if err != nil {
		panic(err)
	}
	return commit
}

// mappedVersion extracts mapped version from current commit message
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
	return strings.TrimSpace(version)
}

// createTag creates new Git tag in repository
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

// createBranch creates new branch pointing to specified commit SHA
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

func (d *DefaultClient) removeBranch(branchName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	_, err := d.client.Git.DeleteRef(ctx, d.owner, d.repo, branchRef(branchName))

	return err
}

func (d *DefaultClient) checkVersion(ver *versions.Versions, cfg config.LLPkgConfig) {
	// 4. Check MappedVersion
	version := d.checkMappedVersion(cfg.Upstream.Package.Name)
	_, mappedVersion := parseMappedVersion(version)

	// 5. Check version is valid
	checkLegacyVersion(ver, cfg, mappedVersion)
}

// CheckPR validates PR changes and returns affected packages
func (d *DefaultClient) CheckPR() []string {
	// build a file path map
	pathMap := map[string][]string{}
	for _, path := range Changes() {
		dir := filepath.Dir(path)
		pathMap[dir] = append(pathMap[dir], filepath.Base(path))
	}
	var allPaths []string

	ver := versions.Read("llpkgstore.json")

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
		d.checkVersion(ver, cfg)

		allPaths = append(allPaths, path)
	}

	// 1. Check there's only one directory in PR
	if len(pathMap) > 1 {
		panic("too many to-be-converted directory")
	}

	// 2. Check config files(llpkg.cfg and llcppg.cfg)
	if len(pathMap) == 0 {
		panic("no valid config files, llpkg.cfg and llcppg.cfg must exist")
	}

	return allPaths
}

// Release handles version tagging and record updates after PR merge
func (d *DefaultClient) Release() {
	// https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#push
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		panic("no GITHUB_SHA found")
	}
	// check it's associated with a pr
	if !d.isAssociatedWithPullRequest(sha) {
		// not a merge commit, skip it.
		panic("not a merge request commit")
	}

	version := d.mappedVersion()
	// skip it when no mapped version is found
	if version == "" {
		panic("no mapped version found in the commit message")
	}

	if hasTag(version) {
		panic("tag has already existed")
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
	ver := versions.Read("llpkgstore.json")
	ver.Write(clib, config.Upstream.Package.Version, mappedVersion)

	// we have finished tagging the commit, safe to remove the branch
	if branchName, isLegacy := isLegacyVersion(); isLegacy {
		d.removeBranch(branchName)
	}
	// move to website in Github Action...
}

// CreateBranchFromLabel creates release branch based on label format
func (d *DefaultClient) CreateBranchFromLabel(labelName string) {
	// design: branch:release-branch.{CLibraryName}/{MappedVersion}
	branchName := strings.TrimPrefix(strings.TrimSpace(labelName), LabelPrefix)
	if branchName == labelName {
		panic("invalid label name format")
	}

	// fast-path: branch exists, can skip.
	if d.hasBranch(branchName) {
		return
	}
	version := strings.TrimPrefix(branchName, BranchPrefix)
	if version == branchName {
		panic("invalid label name format")
	}
	clib, _ := parseMappedVersion(version)
	// slow-path: check the condition if we can create a branch
	//
	// create a branch only when this version is legacy.
	// according to branch maintenance strategy

	// get latest version of the clib
	ver := versions.Read("llpkgstore.json")

	cversions := ver.CVersions(clib)
	if len(cversions) == 0 {
		panic("no clib found")
	}

	if !versions.IsSemver(cversions) {
		panic("c version dones't follow semver, skip maintaining.")
	}

	if err := d.createBranch(branchName, shaFromTag(version)); err != nil {
		panic(err)
	}
}

func (d *DefaultClient) CleanResource() {
	issueEvent := IssueEvent()

	issueNumber := int(issueEvent["number"].(float64))
	regex := regexp.MustCompile(fmt.Sprintf(`(f|F)ix.*#%d`, issueNumber))

	found := false
	for _, commit := range d.allCommits() {
		message := commit.Commit.GetMessage()

		if regex.MatchString(message) &&
			d.isAssociatedWithPullRequest(commit.GetSHA()) {
			found = true
			break
		}
	}

	if !found {
		panic("current issue isn't closed by merged PR.")
	}

	var labelName string

	for _, labels := range issueEvent["labels"].([]map[string]any) {
		label := labels["name"].(string)

		if strings.HasPrefix(label, BranchPrefix) {
			labelName = label
			break
		}
	}

	if labelName == "" {
		panic("current issue hasn't labelled, this should not happen")
	}

	d.removeLabel(labelName)
}
