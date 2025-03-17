package actions

import (
	"os/exec"
	"testing"
)

func TestHasTag(t *testing.T) {
	if hasTag("test1.0.0") {
		t.Error("unexpected tag")
	}
	exec.Command("git", "tag", "test1.0.0").Run()
	if !hasTag("test1.0.0") {
		t.Error("tag doesn't exist")
	}
	ret, _ := exec.Command("git", "tag").CombinedOutput()
	t.Log(string(ret))
	exec.Command("git", "tag", "-d", "test1.0.0").Run()
	if hasTag("test1.0.0") {
		t.Error("unexpected tag")
	}
}
