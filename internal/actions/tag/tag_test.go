package tag

import (
	"os/exec"
	"strings"
	"testing"
)

func TestTag(t *testing.T) {
	tg := From("aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9")
	if tg.Exist() {
		t.Error("unexpected tag")
	}
	if tg.Ref() != "refs/tags/aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9" {
		t.Errorf("unexpected tag ref: want %s got %s", "aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9", tg.Ref())
	}
	exec.Command("git", "tag", "aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9").Run()
	if !tg.Exist() {
		t.Error("tag doesn't exist")
	}
	ret, _ := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if strings.TrimSpace(string(ret)) != tg.SHA() {
		t.Errorf("unexpected tag SHA: want %s got %s", string(ret), tg.SHA())
	}
	ret, _ = exec.Command("git", "tag").CombinedOutput()
	t.Log(string(ret))
	exec.Command("git", "tag", "-d", "aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9").Run()
	if tg.Exist() {
		t.Error("unexpected tag")
	}
}
