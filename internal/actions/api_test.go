package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goplus/llpkgstore/internal/actions/llpkg"
	"github.com/goplus/llpkgstore/internal/actions/mappingtable"
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

func prepareEnv(llpkgConfig, mappingTable []byte) (testDir string, err error) {
	testDir, err = os.MkdirTemp("", "action-test")
	if err != nil {
		return
	}
	err = os.WriteFile(filepath.Join(testDir, "llpkg.cfg"), []byte(llpkgConfig), 0755)
	if err != nil {
		os.RemoveAll(testDir)
		return
	}
	err = os.WriteFile(filepath.Join(testDir, "llpkgstore.json"), mappingTable, 0644)
	if err != nil {
		os.RemoveAll(testDir)
		return
	}

	os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(`module cjson
	go 1.22
	`), 0644)

	os.WriteFile(filepath.Join(testDir, "x.go"), []byte(`package cjson`), 0644)
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

	testMappingTable := `{
		"cjson": {
			"versions" : {
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.2", "v0.1.3"],
				"1.8.18": ["v0.1.0", "v0.1.1"]
			}
		}
	}`

	testDir, err := prepareEnv([]byte(testLLPkgConfig), []byte(testMappingTable))
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(testDir)

	pkg, err := llpkg.NewLLPkg(testDir)
	if err != nil {
		t.Error(err)
		return
	}
	ver := mappingtable.Read(filepath.Join(testDir, "llpkgstore.json"))

	err = actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, pkg, "v0.1.1", legacy)
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

	testMappingTable := `{
		"cjson": {
			"versions" : {
				"1.8.18": ["v0.2.0", "v0.2.1"],
				"1.7.18": ["v0.1.0", "v0.1.1"],
				"1.7.16: ["v1.1.0"]
			}
		}
	}`

	testDir, err := prepareEnv([]byte(testLLPkgConfig), []byte(testMappingTable))
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(testDir)

	pkg, err := llpkg.NewLLPkg(testDir)
	if err != nil {
		t.Error(err)
		return
	}
	ver := mappingtable.Read(filepath.Join(testDir, "llpkgstore.json"))

	err = actionFn("release-branch.cjson/v0.1.1", func(legacy bool) error {
		return checkLegacyVersion(ver, pkg, "v0.1.2", legacy)
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

	testMappingTable := `{
		"cjson": {
			"versions" : {
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.1", "v0.1.2"],
				"1.8.18": ["v0.2.0", "v0.2.1"]
			}
		}
	}`

	testDir, err := prepareEnv([]byte(testLLPkgConfig), []byte(testMappingTable))
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(testDir)

	pkg, err := llpkg.NewLLPkg(testDir)
	if err != nil {
		t.Error(err)
		return
	}
	ver := mappingtable.Read(filepath.Join(testDir, "llpkgstore.json"))

	err = actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, pkg, "v0.3.0", legacy)
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

	testMappingTable := `{
		"cjson": {
			"versions" : {
				"1.8.18": ["v0.2.0", "v0.2.1"],
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.1", "v0.1.2"]
			}
		}
	}`

	testDir, err := prepareEnv([]byte(testLLPkgConfig), []byte(testMappingTable))
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(testDir)

	pkg, err := llpkg.NewLLPkg(testDir)
	if err != nil {
		t.Error(err)
		return
	}
	ver := mappingtable.Read(filepath.Join(testDir, "llpkgstore.json"))

	err = actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, pkg, "v0.0.1", legacy)
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

	testMappingTable := `{
		"cjson": {
			"versions" : {
				"1.7.16": ["v0.1.0"],
				"1.7.18": ["v0.1.2", "v0.1.3"],
				"1.8.18": ["v0.2.0", "v0.2.1"]
			}
		}
	}`

	testDir, err := prepareEnv([]byte(testLLPkgConfig), []byte(testMappingTable))
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(testDir)

	pkg, err := llpkg.NewLLPkg(testDir)
	if err != nil {
		t.Error(err)
		return
	}
	ver := mappingtable.Read(filepath.Join(testDir, "llpkgstore.json"))

	err = actionFn("main", func(legacy bool) error {
		return checkLegacyVersion(ver, pkg, "v0.1.1", legacy)
	})

	if err == nil {
		t.Errorf("unexpected behavior: %v", err)
		return
	}
}
