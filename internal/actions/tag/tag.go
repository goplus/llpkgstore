package tag

import (
	"log"
	"os/exec"
	"strings"
)

type Tag string

func From(tag string) Tag {
	return Tag(strings.TrimSpace(tag))
}

// SHA retrieves commit SHA for given Git tag
// Panics if tag doesn't exist
func (t Tag) SHA() string {
	ret, err := exec.Command("git", "rev-list", "-n", "1", t.String()).CombinedOutput()
	if err != nil {
		log.Panicf("cannot find a tag: %s %s", t, string(ret))
	}
	return strings.TrimSpace(string(ret))
}

// Exist checks if specified Git tag exists in repository
func (t Tag) Exist() bool {
	_, err := exec.Command("git", "rev-parse", t.Ref()).CombinedOutput()
	return err == nil
}

// tagRef constructs full Git tag reference string (e.g. "refs/tags/v1.0.0")
func (t Tag) Ref() string {
	return "refs/tags/" + t.String()
}

func (t Tag) String() string {
	return string(t)
}
