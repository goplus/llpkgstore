package actions

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/env"
	"github.com/goplus/llpkgstore/internal/actions/parser/mappedversion"
	"github.com/goplus/llpkgstore/internal/actions/parser/prefix"
	"github.com/goplus/llpkgstore/internal/actions/tag"
	"github.com/goplus/llpkgstore/internal/actions/versions"
	"github.com/goplus/llpkgstore/internal/file"
	"github.com/goplus/llpkgstore/internal/pc"
)

const _defaultReleaseBranch = "main"

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
func NewDefaultClient() *DefaultClient {
	dc := &DefaultClient{
		client: github.NewClient(nil).WithAuthToken(env.Token()),
	}
	dc.owner, dc.repo = env.Repository()
	return dc
}

// CheckPR validates PR changes and returns affected packages
func (d *DefaultClient) CheckPR() []string {
	// build a file path map
	pathMap := map[string][]string{}
	for _, path := range env.Changes() {
		dir := filepath.Dir(path)
		// initialize the dir
		pathMap[dir] = nil
	}

	var allPaths []string

	ver := versions.Read("llpkgstore.json")

	for path := range pathMap {
		// don't retrieve files from pr changes, consider about maintenance case
		files, _ := os.ReadDir(path)

		if !isValidLLPkg(files) {
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

// Postprocessing handles version tagging and record updates after PR merge
// Creates Git tags, updates version records, and cleans up legacy branches
func (d *DefaultClient) Postprocessing() {
	// https://docs.github.com/en/actions/writing-workflows/choosing-when-your-workflow-runs/events-that-trigger-workflows#push
	sha := env.LatestCommitSHA()
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

	clib, mappedVersion := mappedversion.From(version).MustParse()

	// the pr has merged, so we can read it.
	cfg, err := config.ParseLLPkgConfig(filepath.Join(clib, "llpkg.cfg"))
	must(err)

	// write it to llpkgstore.json
	ver := versions.Read("llpkgstore.json")
	ver.Write(clib, cfg.Upstream.Package.Version, mappedVersion)

	versionTag := tag.From(version)

	if versionTag.Exist() {
		panic("tag has already existed")
	}

	if err := d.createTag(versionTag, sha); err != nil {
		panic(err)
	}

	// create a release
	release := d.createReleaseByTag(version)

	d.uploadArtifactsToRelease(release)

	// we have finished tagging the commit, safe to remove the branch
	if branchName, isLegacy := d.isLegacyVersion(); isLegacy {
		d.removeBranch(branchName)
	}
	// move to website in Github Action...
}

func (d *DefaultClient) Release() {
	version := d.mappedVersion()
	// skip it when no mapped version is found
	if version == "" {
		panic("no mapped version found in the commit message")
	}

	clib, _ := mappedversion.From(version).MustParse()
	// the pr has merged, so we can read it.
	cfg, err := config.ParseLLPkgConfig(filepath.Join(clib, "llpkg.cfg"))
	must(err)

	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	must(err)

	tempDir, _ := os.MkdirTemp("", "llpkg-tool")

	deps, err := uc.Installer.Install(uc.Pkg, tempDir)
	must(err)

	pkgConfigDir := filepath.Join(tempDir, "lib", "pkgconfig")
	// clear exist .pc
	os.RemoveAll(pkgConfigDir)

	err = os.Mkdir(pkgConfigDir, 0777)
	must(err)

	for _, pcName := range deps {
		pcFile := filepath.Join(tempDir, pcName+".pc")
		// generate pc template to lib/pkgconfig
		err = pc.GenerateTemplateFromPC(pcFile, pkgConfigDir, deps)
		must(err)
	}

	// okay, safe to remove old pc
	file.RemovePattern(filepath.Join(tempDir, "*.pc"))
	file.RemovePattern(filepath.Join(tempDir, "*.sh"))

	zipFilename := binaryZip(uc.Pkg.Name)
	zipFilePath, _ := filepath.Abs(zipFilename)

	err = file.Zip(tempDir, zipFilePath)
	must(err)

	// upload to artifacts in GitHub Action
	env.Setenv(env.Env{
		"BIN_PATH":     zipFilePath,
		"BIN_FILENAME": strings.TrimSuffix(zipFilename, ".zip"),
	})
}

// CreateBranchFromLabel creates release branch based on label format
// Follows naming convention: release-branch.<CLibraryName>/<MappedVersion>
func (d *DefaultClient) CreateBranchFromLabel(labelName string) {
	// design: branch:release-branch.{CLibraryName}/{MappedVersion}
	branchName := prefix.NewLabelParser(labelName).MustParse()

	// fast-path: branch exists, can skip.
	if d.hasBranch(branchName) {
		return
	}
	version := prefix.NewBranchParser(branchName).MustParse()

	clib, _ := mappedversion.From(version).MustParse()
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

	err := d.createBranch(branchName, tag.From(version).SHA())
	must(err)
}

// CleanResource removes labels and resources after issue resolution
// Verifies issue closure via PR merge before deletion
func (d *DefaultClient) CleanResource() {
	issueEvent := IssueEvent()

	issueNumber := int(issueEvent["number"].(float64))
	regex := regexp.MustCompile(fmt.Sprintf(`(f|F)ix.*#%d`, issueNumber))

	// 1. check this issue is closed by a PR
	// In Github, close a issue with a commit whose message follows this format
	// fix/Fix* #{IssueNumber}
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

	// 2. find out the branch name from the label
	for _, labels := range issueEvent["labels"].([]map[string]any) {
		label := labels["name"].(string)

		if strings.HasPrefix(label, prefix.BranchPrefix) {
			labelName = label
			break
		}
	}

	if labelName == "" {
		panic("current issue hasn't labelled, this should not happen")
	}

	d.removeLabel(labelName)
}

// hasBranch checks existence of a specific branch in the repository
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
func (d *DefaultClient) associatedWithPullRequest(sha string) []*github.PullRequest {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	pulls, _, err := d.client.PullRequests.ListPullRequestsWithCommit(
		ctx, d.owner, d.repo, sha, &github.ListOptions{},
	)
	must(err)
	return pulls
}

// isAssociatedWithPullRequest checks if commit belongs to a closed pull request
func (d *DefaultClient) isAssociatedWithPullRequest(sha string) bool {
	pulls := d.associatedWithPullRequest(sha)
	// don't use GetMerge, because GetMerge may be a mistake.
	// sometime, when a pull request is merged, GetMerge still returns false.
	// so checking pull request state is more accurate.
	return len(pulls) > 0 &&
		pulls[0].GetState() == "closed"
}

// isLegacyVersion determines if PR targets a legacy branch
func (d *DefaultClient) isLegacyVersion() (branchName string, legacy bool) {
	pullRequest, ok := GitHubEvent()["pull_request"].(map[string]any)
	if !ok {
		// if this actions is not triggered by pull request, fallback to call API.
		pulls := d.associatedWithPullRequest(env.LatestCommitSHA())
		if len(pulls) == 0 {
			panic("this commit is not associated with a pull request, this should not happen")
		}
		branchName = pulls[0].GetBase().GetRef()
	} else {
		// unnecessary to check type, because currentPRCommit has been checked.
		base := pullRequest["base"].(map[string]any)
		branchName = base["ref"].(string)
	}

	legacy = isLegacyBranch(branchName)
	return
}

// currentPRCommit retrieves all commits in the current pull request
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
	must(err)
	return commits
}

// allCommits retrieves all repository commits
func (d *DefaultClient) allCommits() []*github.RepositoryCommit {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	commits, _, err := d.client.Repositories.ListCommits(
		ctx, d.owner, d.repo,
		&github.CommitsListOptions{},
	)
	must(err)
	return commits
}

// removeLabel deletes a label from the repository
func (d *DefaultClient) removeLabel(labelName string) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	// use authorized API to avoid Github RateLimit
	_, err := d.client.Issues.DeleteLabel(
		ctx, d.owner, d.repo, labelName,
	)
	must(err)
}

// checkMappedVersion validates PR contains valid "Release-as" version declaration
func (d *DefaultClient) checkMappedVersion(packageName string) mappedversion.MappedVersion {
	matchMappedVersion := regex(packageName)

	var rawMappedVersion string

	for _, commit := range d.currentPRCommit() {
		message := commit.GetCommit().GetMessage()
		if rawMappedVersion = matchMappedVersion.FindString(message); rawMappedVersion != "" {
			// remove space, of course
			rawMappedVersion = strings.TrimSpace(rawMappedVersion)
			break
		}
	}

	if rawMappedVersion == "" {
		panic("no MappedVersion found in the PR")
	}

	return mappedversion.From(rawMappedVersion)
}

// commitMessage retrieves commit details by SHA
func (d *DefaultClient) commitMessage(sha string) *github.RepositoryCommit {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	commit, _, err := d.client.Repositories.GetCommit(ctx, d.owner, d.repo, sha, &github.ListOptions{})
	must(err)
	return commit
}

// mappedVersion parses the latest commit's mapped version from "Release-as" directive
func (d *DefaultClient) mappedVersion() string {
	// get message
	message := d.commitMessage(env.LatestCommitSHA()).GetCommit().GetMessage()

	// parse the mapped version
	commitVersion := regex(".*").FindString(message)
	// mapped version not found, a normal commit?
	if commitVersion == "" {
		return ""
	}
	version := prefix.NewCommitVersionParser(commitVersion).MustParse()

	return strings.TrimSpace(version)
}

// createTag creates a new Git tag pointing to specific commit
func (d *DefaultClient) createTag(versionTag tag.Tag, sha string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	// tag the commit
	tagRefName := versionTag.Ref()
	_, _, err := d.client.Git.CreateRef(ctx, d.owner, d.repo, &github.Reference{
		Ref: &tagRefName,
		Object: &github.GitObject{
			SHA: &sha,
		},
	})

	return err
}

// createBranch creates a new branch pointing to specific commit
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

func (d *DefaultClient) createReleaseByTag(tag string) *github.RepositoryRelease {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branch := _defaultReleaseBranch

	makeLatest := "true"
	if _, isLegacy := d.isLegacyVersion(); isLegacy {
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
	must(err)

	return release
}

func (d *DefaultClient) uploadToRelease(fileName string, size int64, reader io.Reader, release *github.RepositoryRelease) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("repos/%s/%s/releases/%d/assets?name=%s", d.owner, d.repo, release.GetID(), fileName)

	req, err := d.client.NewUploadRequest(url, reader, size, "application/zip")
	must(err)

	asset := new(github.ReleaseAsset)
	_, err = d.client.Do(ctx, req, asset)
	must(err)
}

func (d *DefaultClient) uploadArtifactToRelease(wg *sync.WaitGroup, artifactID int64, release *github.RepositoryRelease) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer wg.Done()
	defer cancel()

	url, _, err := d.client.Actions.DownloadArtifact(ctx, d.owner, d.repo,
		artifactID, 0)

	must(err)

	httpClient := &http.Client{Timeout: 30 * time.Second}

	resp, err := httpClient.Get(url.String())
	must(err)
	defer resp.Body.Close()

	disposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(disposition)
	must(err)

	fileName, ok := params["filename"]
	if !ok {
		panic("no filename found in Content-Disposition")
	}

	fmt.Printf("Upload %s to %s\n", fileName, release.GetName())

	d.uploadToRelease(fileName, resp.ContentLength, resp.Body, release)
}

func (d *DefaultClient) uploadArtifactsToRelease(release *github.RepositoryRelease) (files []*os.File) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	artifacts, _, err := d.client.Actions.ListWorkflowRunArtifacts(ctx, d.owner, d.repo,
		env.WorkflowID(), &github.ListOptions{})

	must(err)

	if artifacts.GetTotalCount() == 0 {
		panic("no artifact found")
	}

	var wg sync.WaitGroup
	wg.Add(len(artifacts.Artifacts))
	for _, artifact := range artifacts.Artifacts {
		go d.uploadArtifactToRelease(&wg, artifact.GetID(), release)
	}
	wg.Wait()
	return
}

// removeBranch deletes a branch from the repository
func (d *DefaultClient) removeBranch(branchName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	_, err := d.client.Git.DeleteRef(ctx, d.owner, d.repo, branchRef(branchName))

	return err
}

// checkVersion performs version validation and configuration checks
func (d *DefaultClient) checkVersion(ver *versions.Versions, cfg config.LLPkgConfig) {
	// 4. Check MappedVersion
	version := d.checkMappedVersion(cfg.Upstream.Package.Name)
	_, mappedVersion := version.MustParse()
	// 5. Check version is valid
	_, isLegacy := d.isLegacyVersion()
	checkLegacyVersion(ver, cfg, mappedVersion, isLegacy)
}
