package prefix

import "testing"

func TestLabelPrefix(t *testing.T) {
	t.Run("label-valid", func(t *testing.T) {
		p := NewLabelParser("branch:release-branch.").MustParse()
		if p != "release-branch." {
			t.Errorf("unexpected result: want: %s got %s", "release-branch.", p)
		}
	})
	t.Run("inabel-valid", func(t *testing.T) {
		_, err := NewLabelParser("release-branch.").Parse()
		if err == nil {
			t.Error("unexpected behavior: no error")
		}
	})
}

func TestBranchPrefix(t *testing.T) {
	t.Run("branch-valid", func(t *testing.T) {
		p := NewBranchParser("release-branch.cjson/v1.0.0").MustParse()
		if p != "cjson/v1.0.0" {
			t.Errorf("unexpected result: want: %s got %s", "cjson/v1.0.0", p)
		}
	})
	t.Run("branch-invalid", func(t *testing.T) {
		_, err := NewBranchParser("release-branch").Parse()
		if err == nil {
			t.Error("unexpected behavior: no error")
		}
	})
}

func TestCommitVersionPrefix(t *testing.T) {
	t.Run("commitversion-valid", func(t *testing.T) {
		p := NewCommitVersionParser("Release-as: cjson/v1.0.0").MustParse()
		if p != "cjson/v1.0.0" {
			t.Errorf("unexpected result: want: %s got %s", "cjson/v1.0.0", p)
		}
	})
	t.Run("commitversion-invalid", func(t *testing.T) {
		_, err := NewCommitVersionParser("Release: cjson/v1.0.0").Parse()
		if err == nil {
			t.Error("unexpected behavior: no error")
		}
	})
}
