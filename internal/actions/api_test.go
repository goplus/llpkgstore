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

func recoverFn(branchName string, fn func(legacy bool)) (ret any) {
	defer func() {
		ret = recover()
	}()
	fn(strings.HasPrefix(branchName, BranchPrefix))
	return
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
			"versions" : [{
				"c": "1.8.18",
				"go": ["v0.1.0", "v0.1.1"]
			},{
				"c": "1.7.18",
				"go": ["v0.1.2", "v0.1.3"]
			},
			{
				"c": "1.7.16",
				"go": ["v0.1.0"]
			}]
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := recoverFn("main", func(legacy bool) {
		checkLegacyVersion(ver, cfg, "v0.1.1", legacy)
	})
	_, ok := err.(string)
	isValid := ok && err != ""

	if !isValid {
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
			"versions" : [{
				"c": "1.8.18",
				"go": ["v0.2.0", "v0.2.1"]
			},{
				"c": "1.7.18",
				"go": ["v0.1.0", "v0.1.1"]
			},
			{
				"c": "1.7.16",
				"go": ["v1.1.0"]
			}]
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := recoverFn("release-branch.cjson/v0.1.1", func(legacy bool) {
		checkLegacyVersion(ver, cfg, "v0.1.2", legacy)
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
			"versions" : [{
				"c": "1.8.18",
				"go": ["v0.2.0", "v0.2.1"]
			},{
				"c": "1.7.18",
				"go": ["v0.1.1", "v0.1.2"]
			},
			{
				"c": "1.7.16",
				"go": ["v0.1.0"]
			}]
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := recoverFn("main", func(legacy bool) {
		checkLegacyVersion(ver, cfg, "v0.3.0", legacy)
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
			"versions" : [{
				"c": "1.8.18",
				"go": ["v0.2.0", "v0.2.1"]
			},{
				"c": "1.7.18",
				"go": ["v0.1.1", "v0.1.2"]
			},
			{
				"c": "1.7.16",
				"go": ["v0.1.0"]
			}]
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := recoverFn("main", func(legacy bool) {
		checkLegacyVersion(ver, cfg, "v0.0.1", legacy)
	})
	_, ok := err.(string)
	isValid := ok && err != ""

	if !isValid {
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
			"versions" : [{
				"c": "1.8.18",
				"go": ["v0.2.0", "v0.2.1"]
			},{
				"c": "1.7.18",
				"go": ["v0.1.2", "v0.1.3"]
			},
			{
				"c": "1.7.16",
				"go": ["v0.1.0"]
			}]
		}
	}`)

	os.WriteFile(".llpkgstore.json", []byte(b), 0755)
	defer os.Remove(".llpkgstore.json")

	cfg, _ := config.ParseLLPkgConfig(".llpkg.cfg")
	ver := versions.Read(".llpkgstore.json")

	err := recoverFn("main", func(legacy bool) {
		checkLegacyVersion(ver, cfg, "v0.1.1", legacy)
	})
	_, ok := err.(string)
	isValid := ok && err != ""

	if !isValid {
		t.Errorf("unexpected behavior: %v", err)
		return
	}
}
