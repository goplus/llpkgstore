package env

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// must panics if the error is non-nil, halting execution
func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Env map[string]string

// String converts environment variables map to newline-separated key=value pairs for GitHub Actions
func (e Env) String() string {
	var env []string

	for name, value := range e {
		env = append(env, fmt.Sprintf("%s=%s", name, value))
	}

	sort.Strings(env)
	return strings.Join(env, "\n")
}

// Setenv writes environment variables to GITHUB_ENV for GitHub Actions consumption
func Setenv(envm Env) {
	env, err := os.OpenFile(os.Getenv("GITHUB_ENV"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	// should never happen,
	// it means current runtime is not Github actions if there's any errors
	must(err)

	env.WriteString(envm.String())

	// make sure we write it to the GITHUB_ENV
	env.Close()
}

// SetOutput writes workflow outputs to GITHUB_OUTPUT for GitHub Actions
func SetOutput(envm Env) {
	env, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	must(err)

	env.WriteString(envm.String())

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
// Repository extracts GitHub repository owner and name from GITHUB_REPOSITORY
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
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("no GITHUB_TOKEN")
	}
	return token
}

// LatestCommitSHA returns the current commit SHA from GITHUB_SHA environment variable
func LatestCommitSHA() string {
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		panic("no GITHUB_SHA found")
	}
	return sha
}

func WorkflowID() int64 {
	runId := os.Getenv("GITHUB_RUN_ID")
	if runId == "" {
		panic("no GITHUB_RUN_ID found")
	}
	id, err := strconv.ParseInt(runId, 10, 64)
	must(err)
	return id
}
