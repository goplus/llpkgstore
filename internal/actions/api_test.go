package actions

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/versions"
)

func TestHasTag(t *testing.T) {
	if hasTag("aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9") {
		t.Error("unexpected tag")
	}
	exec.Command("git", "tag", "aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9").Run()
	if !hasTag("aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9") {
		t.Error("tag doesn't exist")
	}
	ret, _ := exec.Command("git", "tag").CombinedOutput()
	t.Log(string(ret))
	exec.Command("git", "tag", "-d", "aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9").Run()
	if hasTag("aaaaaaaaaaa1.1.4.5.1.4.1.9.1.9") {
		t.Error("unexpected tag")
	}
}

func actionFn(branchName string, fn func(legacy bool) error) error {
	return fn(strings.HasPrefix(branchName, BranchPrefix))
}

func TestLegacyVersion1(t *testing.T) {
	testLLPkgConfig := `{
		"upstream": {
		  "package": {
			"name": "cjson",
			"version": "1.7.17"
		  }
		}
	  }`
	os.WriteFile(".llpkg.cfg", []byte(testLLPkgConfig), 0755)
	defer os.Remove(".llpkg.cfg")

	b := []byte(`{
		"cjson": {
			"versions" : {
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.2", "v0.1.3"],
				"1.8.18": ["v0.1.0", "v0.1.1"]
			}
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, cfg, "v0.1.1", legacy)
	})

	if err == nil {
		t.Errorf("unexpected behavior: %v", err)
		return
	}

}

func TestLegacyVersion2(t *testing.T) {
	testLLPkgConfig := `{
		"upstream": {
		  "package": {
			"name": "cjson",
			"version": "1.7.19"
		  }
		}
	  }`
	os.WriteFile(".llpkg.cfg", []byte(testLLPkgConfig), 0755)
	defer os.Remove(".llpkg.cfg")

	b := []byte(`{
		"cjson": {
			"versions" : {
				"1.8.18": ["v0.2.0", "v0.2.1"],
				"1.7.18": ["v0.1.0", "v0.1.1"],
				"1.7.16: ["v1.1.0"]
			}
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := actionFn("release-branch.cjson/v0.1.1", func(legacy bool) error {
		return checkLegacyVersion(ver, cfg, "v0.1.2", legacy)
	})
	isValid := err == nil

	if !isValid {
		t.Errorf("unexpected behavior: %v", err)
		return
	}
}

func TestLegacyVersion3(t *testing.T) {
	testLLPkgConfig := `{
		"upstream": {
		  "package": {
			"name": "cjson",
			"version": "1.9.1"
		  }
		}
	  }`
	os.WriteFile(".llpkg.cfg", []byte(testLLPkgConfig), 0755)
	defer os.Remove(".llpkg.cfg")

	b := []byte(`{
		"cjson": {
			"versions" : {
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.1", "v0.1.2"],
				"1.8.18": ["v0.2.0", "v0.2.1"]
			}
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, cfg, "v0.3.0", legacy)
	})
	isValid := err == nil

	if !isValid {
		t.Errorf("unexpected behavior: %v", err)
		return
	}
}

func TestLegacyVersion4(t *testing.T) {
	testLLPkgConfig := `{
		"upstream": {
		  "package": {
			"name": "cjson",
			"version": "1.9.1"
		  }
		}
	  }`
	os.WriteFile(".llpkg.cfg", []byte(testLLPkgConfig), 0755)
	defer os.Remove(".llpkg.cfg")

	b := []byte(`{
		"cjson": {
			"versions" : {
				"1.8.18": ["v0.2.0", "v0.2.1"],
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.1", "v0.1.2"]
			}
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, cfg, "v0.0.1", legacy)
	})

	if err == nil {
		t.Errorf("unexpected behavior: %v", err)
		return
	}
}

func TestLegacyVersion5(t *testing.T) {
	testLLPkgConfig := `{
		"upstream": {
		  "package": {
			"name": "cjson",
			"version": "1.7.19"
		  }
		}
	  }`
	os.WriteFile(".llpkg.cfg", []byte(testLLPkgConfig), 0755)
	defer os.Remove(".llpkg.cfg")

	b := []byte(`{
		"cjson": {
			"versions" : {
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.2", "v0.1.3"],
				"1.8.18": ["v0.2.0", "v0.2.1"]
			}
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, cfg, "v0.1.1", legacy)
	})

	if err == nil {
		t.Errorf("unexpected behavior: %v", err)
		return
	}
}
