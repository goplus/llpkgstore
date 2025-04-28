package actions

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goplus/llpkgstore/internal/actions/env"
	"github.com/goplus/llpkgstore/internal/actions/llpkg"
	"github.com/goplus/llpkgstore/internal/actions/mappingtable"
	"github.com/goplus/llpkgstore/internal/actions/versions"
)

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

	// the pr has merged, so we can read it.
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

	// the pr has merged, so we can read it.
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
