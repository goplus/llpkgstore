package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/goplus/llpkgstore/config"
)

const (
	regexString = `Release-as:\s%s/v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

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

func (d *DefaultClient) hasBranch(branchName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	branch, resp, err := d.client.Repositories.GetBranch(
		ctx, d.owner, d.repo, branchName, 0,
	)
	exists := err == nil &&
		branch != nil &&
		resp.StatusCode == http.StatusOK
	return exists, err
}

func (d *DefaultClient) hasTag(tag string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	tags, _, err := d.client.Repositories.ListTags(
		ctx, d.owner, d.repo, &github.ListOptions{},
	)
	if err != nil {
		return false, err
	}
	found := false
	for _, current := range tags {
		if current.GetName() == tag {
			found = true
			break
		}
	}
	return found, nil
}

func (d *DefaultClient) isAssociatedWithPullRequest(sha string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	pulls, _, err := d.client.PullRequests.ListPullRequestsWithCommit(
		ctx, d.owner, d.repo, sha, &github.ListOptions{},
	)

	return len(pulls) > 0 &&
		pulls[0].GetMerged(), err
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
	ok, err := d.isAssociatedWithPullRequest(sha)
	if err != nil {
		panic(err)
	}
	// not a merge commit, skip it.
	if !ok {
		return
	}
	version := mappedVersion()
	ok, err = d.hasTag(version)
	if err != nil {
		panic(err)
	}
	// has tag already, skip it.
	if ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	// tag the commit
	tagRef := "refs/tags/" + version
	_, _, err = d.client.Git.CreateRef(ctx, d.owner, d.repo, &github.Reference{
		Ref: &tagRef,
		Object: &github.GitObject{
			SHA: &sha,
		},
	})

	if err != nil {
		panic(err)
	}

	// TODO: write it to llpkgstore.json
}
