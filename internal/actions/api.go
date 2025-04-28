// Package actions contains GitHub Actions helper functions for version management and repository operations.
package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/goplus/llpkgstore/internal/actions/env"
	"github.com/goplus/llpkgstore/internal/actions/llpkg"
	"github.com/goplus/llpkgstore/internal/actions/mappingtable"
	"github.com/goplus/llpkgstore/internal/actions/versions"
	"golang.org/x/sync/errgroup"
)

const (
	LabelPrefix         = "branch:"
	BranchPrefix        = "release-branch."
	MappedVersionPrefix = "Release-as: "

	defaultReleaseBranch = "main"
	regexString          = `Release-as:\s%s/v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
)

// compileRawCommitVersionRegex compiles a regular expression pattern to detect "Release-as" directives in commit messages
func compileRawCommitVersionRegex(packageName llpkg.PackageName) *regexp.Regexp {
	// format: Release-as: clib/semver(with v prefix)
	// Must have one space in the end of Release-as:
	return regexp.MustCompile(fmt.Sprintf(regexString, packageName.String()))
}

func binaryZip(packageName string) string {
	return fmt.Sprintf("%s_%s.zip", packageName, currentSuffix)
}

// DefaultClient provides GitHub API client capabilities with authentication for Actions workflows
type DefaultClient struct {
	// repo: Target repository name
	// owner: Repository owner organization/user
	// client: Authenticated GitHub API client instance
	repo   string
	owner  string
	client *github.Client
}

// NewDefaultClient initializes a new GitHub API client with authentication and repository configuration
// Uses:
//   - GitHub token from environment
//   - Repository info from GITHUB_REPOSITORY context
//
// Returns:
//
//	*DefaultClient: Configured client instance
func NewDefaultClient() (*DefaultClient, error) {
	token, err := env.Token()
	if err != nil {
		return nil, err
	}
	owner, repo, err := env.Repository()
	if err != nil {
		return nil, err
	}
	dc := &DefaultClient{
		owner: owner, repo: repo,
		client: github.NewClient(nil).WithAuthToken(token),
	}
	return dc, nil
}

// hasBranch checks existence of a specific branch in the repository
// Parameters:
//
//	branchName: Name of the branch to check
//
// Returns:
//
//	bool: True if branch exists
func (d *DefaultClient) hasBranch(branchName string) bool {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branch, resp, err := d.client.Repositories.GetBranch(
		ctx, d.owner, d.repo, branchName, 0,
	)

	return err == nil && branch != nil &&
		resp.StatusCode == http.StatusOK
}

// associatedWithPullRequest finds all pull requests containing the specified commit
// Parameters:
//
//	sha: Commit hash to search for
//
// Returns:
//
//	[]*github.PullRequest: List of associated pull requests
func (d *DefaultClient) associatedWithPullRequest(sha string) ([]*github.PullRequest, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	pulls, _, err := d.client.PullRequests.ListPullRequestsWithCommit(
		ctx, d.owner, d.repo, sha, &github.ListOptions{},
	)
	if err != nil {
		return nil, wrapActionError(err)
	}
	return pulls, nil
}

// isAssociatedWithPullRequest checks if commit belongs to a closed pull request
// Parameters:
//
//	sha: Commit hash to check
//
// Returns:
//
//	bool: True if part of closed PR
func (d *DefaultClient) isAssociatedWithPullRequest(sha string) bool {
	pulls, _ := d.associatedWithPullRequest(sha)
	// don't use GetMerge, because GetMerge may be a mistake.
	// sometime, when a pull request is merged, GetMerge still returns false.
	// so checking pull request state is more accurate.
	return len(pulls) > 0 &&
		pulls[0].GetState() == "closed"
}

// isLegacyVersion determines if PR targets a legacy branch
// Returns:
//
//	branchName: Base branch name
//	legacy: True if branch starts with "release-branch."
func (d *DefaultClient) isLegacyVersion() (branchName string, legacy bool, err error) {
	event, err := GitHubEvent()
	if err != nil {
		return
	}
	pullRequest, ok := event["pull_request"].(map[string]any)
	var refName string
	if !ok {
		var sha string
		var pulls []*github.PullRequest
		sha, err = env.LatestCommitSHA()
		if err != nil {
			return
		}
		// if this actions is not triggered by pull request, fallback to call API.
		pulls, err = d.associatedWithPullRequest(sha)
		if err != nil {
			return
		}
		refName = pulls[0].GetBase().GetRef()
	} else {
		// unnecessary to check type, because currentPRCommit has been checked.
		base := pullRequest["base"].(map[string]any)
		refName = base["ref"].(string)
	}

	legacy = strings.HasPrefix(refName, BranchPrefix)
	branchName = refName
	return
}

// currentPRCommit retrieves all commits in the current pull request
// Returns:
//
//	[]*github.RepositoryCommit: List of PR commits
func (d *DefaultClient) currentPRCommit() ([]*github.RepositoryCommit, error) {
	pullRequest, err := PullRequestEvent()
	if err != nil {
		return nil, err
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
		return nil, wrapActionError(err)
	}
	return commits, nil
}

// allCommits retrieves all repository commits
// Returns:
//
//	[]*github.RepositoryCommit: List of all commits
func (d *DefaultClient) allCommits() ([]*github.RepositoryCommit, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	commits, _, err := d.client.Repositories.ListCommits(
		ctx, d.owner, d.repo,
		&github.CommitsListOptions{},
	)
	if err != nil {
		return nil, wrapActionError(err)
	}
	return commits, nil
}

// removeLabel deletes a label from the repository
// Parameters:
//
//	labelName: Name of the label to remove
func (d *DefaultClient) removeLabel(labelName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	_, err := d.client.Issues.DeleteLabel(
		ctx, d.owner, d.repo, labelName,
	)
	return wrapActionError(err)
}

// checkMappedVersion validates PR contains valid "Release-as" version declaration
// Parameters:
//
//	packageName: Target package name for version mapping
//
// Returns:
//
//	string: Validated mapped version string
//
// Panics:
//
//	If no valid version found in PR commits
func (d *DefaultClient) checkMappedVersion(pkg *llpkg.LLPkg) (mappedVersion string, err error) {
	matchMappedVersion := compileRawCommitVersionRegex(pkg.Name())

	allCommits, err := d.currentPRCommit()
	if err != nil {
		return
	}
	for _, commit := range allCommits {
		message := commit.GetCommit().GetMessage()
		if mappedVersion = matchMappedVersion.FindString(message); mappedVersion != "" {
			// remove space, of course
			mappedVersion = strings.TrimSpace(mappedVersion)
			break
		}
	}

	if mappedVersion == "" {
		err = ErrNoMappedVersion
	}
	return
}

// commitMessage retrieves commit details by SHA
// Parameters:
//
//	sha: Commit hash to retrieve
//
// Returns:
//
//	*github.RepositoryCommit: Commit details object
func (d *DefaultClient) commitMessage(sha string) (*github.RepositoryCommit, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	commit, _, err := d.client.Repositories.GetCommit(ctx, d.owner, d.repo, sha, &github.ListOptions{})
	if err != nil {
		return nil, wrapActionError(err)
	}
	return commit, nil
}

// mappedVersion parses the latest commit's mapped version from "Release-as" directive
// Returns:
//
//	string: Parsed version string or empty if not found
//
// Panics:
//
//	If version format is invalid
func (d *DefaultClient) mappedVersion() (string, error) {
	sha, err := env.LatestCommitSHA()
	if err != nil {
		return "", err
	}
	// get message
	commit, err := d.commitMessage(sha)
	if err != nil {
		return "", err
	}

	message := commit.GetCommit().GetMessage()

	// parse the mapped version
	mappedVersion := compileRawCommitVersionRegex(llpkg.PackageName(".*")).FindString(message)
	// mapped version not found, a normal commit?
	if mappedVersion == "" {
		return "", ErrNoMappedVersion
	}
	version := strings.TrimPrefix(mappedVersion, MappedVersionPrefix)
	if version == mappedVersion {
		return "", fmt.Errorf("actions: invalid format")
	}
	return strings.TrimSpace(version), nil
}

// createTag creates a new Git tag pointing to specific commit
// Parameters:
//
//	tag: Tag name (e.g. "v1.2.3")
//	sha: Target commit hash
//
// Returns:
//
//	error: Error during tag creation
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

	return wrapActionError(err)
}

// createBranch creates a new branch pointing to specific commit
// Parameters:
//
//	branchName: New branch name
//	sha: Target commit hash
//
// Returns:
//
//	error: Error during branch creation
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

	return wrapActionError(err)
}

func (d *DefaultClient) createReleaseByTag(tag string) (*github.RepositoryRelease, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branch := defaultReleaseBranch

	_, isLegacy, err := d.isLegacyVersion()
	if err != nil {
		return nil, err
	}

	makeLatest := "true"
	if isLegacy {
		makeLatest = "legacy"
	}
	generateRelease := true

	release, _, err := d.client.Repositories.CreateRelease(ctx, d.owner, d.repo, &github.RepositoryRelease{
		TagName:              &tag,
		TargetCommitish:      &branch,
		Name:                 &tag,
		MakeLatest:           &makeLatest,
		GenerateReleaseNotes: &generateRelease,
	})
	if err != nil {
		return nil, wrapActionError(err)
	}

	return release, nil
}

func (d *DefaultClient) uploadToRelease(
	fileName string,
	size int64,
	reader io.Reader,
	release *github.RepositoryRelease,
) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("repos/%s/%s/releases/%d/assets?name=%s", d.owner, d.repo, release.GetID(), fileName)

	req, err := d.client.NewUploadRequest(url, reader, size, "application/zip")
	if err != nil {
		return wrapActionError(err)
	}

	asset := new(github.ReleaseAsset)
	_, err = d.client.Do(ctx, req, asset)
	if err != nil {
		return wrapActionError(err)
	}
	return nil
}

func (d *DefaultClient) uploadArtifact(artifactID int64, release *github.RepositoryRelease) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	url, _, err := d.client.Actions.DownloadArtifact(ctx, d.owner, d.repo,
		artifactID, 0)

	if err != nil {
		return wrapActionError(err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	resp, err := httpClient.Get(url.String())
	if err != nil {
		return wrapActionError(err)
	}
	defer resp.Body.Close()

	disposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		return wrapActionError(err)
	}

	fileName, ok := params["filename"]
	if !ok {
		return errors.New("actions: no filename found in Content-Disposition")
	}

	fmt.Printf("Upload %s to %s\n", fileName, release.GetName())

	return d.uploadToRelease(fileName, resp.ContentLength, resp.Body, release)
}

func (d *DefaultClient) uploadArtifactsToRelease(release *github.RepositoryRelease) (files []*os.File, err error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	id, err := env.WorkflowRunID()
	artifacts, _, err := d.client.Actions.ListWorkflowRunArtifacts(ctx, d.owner, d.repo,
		id, &github.ListOptions{})

	if err != nil {
		return nil, wrapActionError(err)
	}

	if artifacts.GetTotalCount() == 0 {
		return nil, errors.New("actions: no artifact found")
	}

	errGroup, _ := errgroup.WithContext(context.TODO())

	for _, artifact := range artifacts.Artifacts {
		// make a copy to avoid for loop bug
		artifactID := artifact.GetID()

		errGroup.Go(func() error {
			return d.uploadArtifact(artifactID, release)
		})
	}

	return nil, errGroup.Wait()
}

// removeBranch deletes a branch from the repository
// Parameters:
//
//	branchName: Name of the branch to delete
//
// Returns:
//
//	error: Error during branch deletion
func (d *DefaultClient) removeBranch(branchName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	_, err := d.client.Git.DeleteRef(ctx, d.owner, d.repo, branchRef(branchName))

	return wrapActionError(err)
}

// checkVersion performs version validation and configuration checks
// Parameters:
//
//	ver: Version store object
//	cfg: Package configuration
func (d *DefaultClient) checkVersion(ver *mappingtable.Versions, pkg *llpkg.LLPkg) error {
	// 4. Check MappedVersion
	version, err := d.checkMappedVersion(pkg)
	if err != nil {
		return err
	}
	_, mappedVersion, err := parseMappedVersion(version)
	if err != nil {
		return err
	}

	// 5. Check version is valid
	_, isLegacy, err := d.isLegacyVersion()
	if err != nil {
		return err
	}
	return checkLegacyVersion(ver, pkg, mappedVersion, isLegacy)
}

// CheckPR validates PR changes and returns affected packages
// Returns:
//
//	[]string: List of affected package paths
func (d *DefaultClient) CheckPR() ([]string, error) {
	// build a file path map
	pathMap := map[string][]string{}
	changedFilePaths, err := env.Changes()
	if err != nil {
		return nil, err
	}
	for _, path := range changedFilePaths {
		dir := filepath.Dir(path)
		// initialize the dir
		pathMap[dir] = nil
	}

	var allPaths []string

	ver := mappingtable.Read("llpkgstore.json")

	for path := range pathMap {
		if !isLLPkgRoot(path) {
			delete(pathMap, path)
			continue
		}
		pkg, err := llpkg.NewLLPkg(path)
		if err != nil {
			return nil, err
		}
		err = d.checkVersion(ver, pkg)
		if err != nil {
			return nil, err
		}
		allPaths = append(allPaths, path)
	}

	// 2. Check config files(llpkg.cfg and llcppg.cfg)
	if len(pathMap) == 0 {
		return nil, fmt.Errorf("actions: no valid config files, llpkg.cfg and llcppg.cfg must exist")
	}

	return allPaths, nil
}

// Postprocessing handles version tagging and record updates after PR merge
// Creates Git tags, updates version records, and cleans up legacy branches
func (d *DefaultClient) Postprocessing() error {
	// https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#push
	sha, err := env.LatestCommitSHA()
	if err != nil {
		return err
	}
	// check it's associated with a pr
	if !d.isAssociatedWithPullRequest(sha) {
		// not a merge commit, skip it.
		return fmt.Errorf("actions: not a merge request commit")
	}

	version, err := d.mappedVersion()
	if err != nil {
		return err
	}

	clib, mappedVersion, err := parseMappedVersion(version)
	if err != nil {
		return err
	}

	pkg, err := llpkg.FromPackageName(clib)
	if err != nil {
		return err
	}

	// write it to llpkgstore.json
	ver := mappingtable.Read("llpkgstore.json")
	ver.Write(pkg.ClibName(), pkg.ClibVersion(), mappedVersion)

	if hasTag(version) {
		return fmt.Errorf("actions: tag has already existed")
	}

	if err := d.createTag(version, sha); err != nil {
		return err
	}

	// create a release
	release, err := d.createReleaseByTag(version)
	if err != nil {
		return err
	}

	_, err = d.uploadArtifactsToRelease(release)
	if err != nil {
		return err
	}

	// we have finished tagging the commit, safe to remove the branch
	branchName, isLegacy, err := d.isLegacyVersion()
	if err != nil {
		return err
	}
	if isLegacy {
		err = d.removeBranch(branchName)
	}
	return err
	// move to website in Github Action...
}

// Release must be called before Postprocessing
func (d *DefaultClient) Release() error {
	version, err := d.mappedVersion()
	if err != nil {
		return err
	}

	clibName, _, err := parseMappedVersion(version)
	if err != nil {
		return err
	}
	pkg, err := llpkg.FromPackageName(clibName)
	if err != nil {
		return err
	}
	uc, err := pkg.Upstream()
	if err != nil {
		return err
	}

	zipFilename, zipFilePath, err := BuildBinaryZip(uc)
	if err != nil {
		return err
	}

	// upload to artifacts in GitHub Action
	// https://github.com/goplus/llpkg/pull/50/files#diff-95373be0ab51a56a2200c8c07981d82e81569f2cd1e4e2946e2002bb66de766fR56-R60
	return env.Setenv(env.Env{
		"BIN_PATH":     zipFilePath,
		"BIN_FILENAME": strings.TrimSuffix(zipFilename, ".zip"),
	})
}

// CreateBranchFromLabel creates release branch based on label format
// Follows naming convention: release-branch.<CLibraryName>/<MappedVersion>
func (d *DefaultClient) CreateBranchFromLabel(labelName string) error {
	// design: branch:release-branch.{CLibraryName}/{MappedVersion}
	branchName := strings.TrimPrefix(strings.TrimSpace(labelName), LabelPrefix)
	if branchName == labelName {
		return fmt.Errorf("actions: invalid label name format")
	}

	// fast-path: branch exists, can skip.
	if d.hasBranch(branchName) {
		return nil
	}
	version := strings.TrimPrefix(branchName, BranchPrefix)
	if version == branchName {
		return fmt.Errorf("actions: invalid label name format")
	}
	clibName, _, err := parseMappedVersion(version)
	if err != nil {
		return err
	}
	pkg, err := llpkg.FromPackageName(clibName)
	if err != nil {
		return err
	}
	// slow-path: check the condition if we can create a branch
	//
	// create a branch only when this version is legacy.
	// according to branch maintenance strategy

	// get latest version of the clib
	ver := mappingtable.Read("llpkgstore.json")

	cversions := ver.CVersions(pkg.ClibName())
	if len(cversions) == 0 {
		return fmt.Errorf("actions: no clib found")
	}

	if !versions.IsSemver(cversions) {
		return fmt.Errorf("actions: c version dones't follow semver, skip maintaining")
	}

	return d.createBranch(branchName, shaFromTag(version))
}

// CleanResource removes labels and resources after issue resolution
// Verifies issue closure via PR merge before deletion
func (d *DefaultClient) CleanResource() error {
	issueEvent, err := IssueEvent()
	if err != nil {
		return err
	}

	issueNumber := int(issueEvent["number"].(float64))
	regex := regexp.MustCompile(fmt.Sprintf(`(f|F)ix.*#%d`, issueNumber))

	// 1. check this issue is closed by a PR
	// In Github, close a issue with a commit whose message follows this format
	// fix/Fix* #{IssueNumber}
	found := false
	allCommits, err := d.allCommits()
	if err != nil {
		return err
	}
	for _, commit := range allCommits {
		message := commit.Commit.GetMessage()

		if regex.MatchString(message) &&
			d.isAssociatedWithPullRequest(commit.GetSHA()) {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("actions: current issue isn't closed by merged PR")
	}

	var labelName string

	// 2. find out the branch name from the label
	for _, labels := range issueEvent["labels"].([]map[string]any) {
		label := labels["name"].(string)

		if strings.HasPrefix(label, BranchPrefix) {
			labelName = label
			break
		}
	}

	if labelName == "" {
		return fmt.Errorf("current issue hasn't labelled, this should not happen")
	}

	return d.removeLabel(labelName)
}
