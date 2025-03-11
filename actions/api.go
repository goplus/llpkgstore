package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/goplus/llpkgstore/config"
)

// format: Release-as: clib/semver(with v prefix)
// Must have one space in the end of Release-as:
var matchMappedVersion = regexp.MustCompile(`Release-as:\s.*/v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

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

func (d *DefaultClient) HasBranch(branchName string) (bool, error) {
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

func (d *DefaultClient) IsAssociatedWithPullRequest(sha string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	pulls, _, err := d.client.PullRequests.ListPullRequestsWithCommit(
		ctx, d.owner, d.repo, sha, &github.ListOptions{},
	)

	return len(pulls) > 0 &&
		pulls[0].GetMerged(), err
}

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

func isValidLlpkg(files []string) bool {
	fileMap := make(map[string]struct{}, len(files))

	for _, file := range files {
		fileMap[file] = struct{}{}
	}
	_, hasLlpkg := fileMap["llpkg.cfg"]
	_, hasLlcppg := fileMap["llcppg.cfg"]
	return hasLlcppg && hasLlpkg
}

func (d *DefaultClient) CheckPR() {
	// build a file path map
	pathMap := map[string][]string{}
	for _, path := range Changes() {
		dir := filepath.Dir(path)
		pathMap[dir] = append(pathMap[dir], filepath.Base(path))
	}

	for path, files := range pathMap {
		if !isValidLlpkg(files) {
			delete(pathMap, path)
			continue
		}
		// 3. Check directory name
		llpkgFile := filepath.Join(path, "llpkg.cfg")
		cfg, err := config.ParseLLpkgConfig(llpkgFile)
		if err != nil {
			panic(err)
		}
		// in our design, directory name should equal to the package name,
		// which means it's not required to be equal.
		//
		// However, in current stage, if this is not equal, conan may panic,
		// to aovid unexpected behavior, we assert it's equal temporarily.
		// this logic may be changed in the future.
		packageName := strings.TrimSpace(cfg.UpstreamConfig.PackageConfig.Name)
		if packageName != path {
			panic("directory name is not equal to package name in llpkg.cfg")
		}
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
