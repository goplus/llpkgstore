package llcppg

import (
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/hashutils"
	"golang.org/x/mod/modfile"
)

const (
	testLLPkgConfig = `{
  "upstream": {
    "package": {
      "name": "cjson",
      "version": "1.7.17"
    }
  }
}`
	testLlcppgConfig = `{
    "name": "cjson",
    "cflags": "$(pkg-config --cflags cjson)",
    "libs": "$(pkg-config --libs cjson)",
    "include": [
            "cjson/cJSON.h"
    ],
    "deps": null,
    "trimPrefixes": [],
    "cplusplus": false
}`
)

func checkGoMod(t *testing.T, file string) {
	b, _ := os.ReadFile(file)
	f, _ := modfile.Parse(file, b, nil)
	if f.Go.Version != "1.20" {
		t.Errorf("unexpected version: got: %s", f.Go.Version)
	}
	if f.Module.Mod.Path != goplusRepo+"cjson" {
		t.Errorf("unexpected module path: got: %s", f.Module.Mod.Path)

	}
}

func TestHash(t *testing.T) {
	canHashFn := func(fileName string) bool {
		return fileName == "gg.test" || fileName == "ggg.test"
	}

	m, err := hashutils.Dir("testfind2", canHashFn)
	if err != nil {
		t.Error(err)
		return
	}
	expectedHash1, _ := hex.DecodeString("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3")
	if !reflect.DeepEqual(m, map[string][]byte{
		"ggg.test": expectedHash1,
	}) {
		t.Errorf("unexpected hash result: want: %v got: %v", map[string][]byte{
			"ggg.test": expectedHash1,
		}, m)
		return
	}
	m2, err := hashutils.Dir("testfind2/testfind", canHashFn)
	if err != nil {
		t.Error(err)
		return
	}
	expectedHash2, _ := hex.DecodeString("a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3")
	expectedHash3, _ := hex.DecodeString("b3a8e0e1f9ab1bfe3a36f231f676f78bb30a519d2b21e6c530c0eee8ebb4a5d0")
	if !reflect.DeepEqual(m2, map[string][]byte{
		"ggg.test": expectedHash2,
		"gg.test":  expectedHash3,
	}) {
		t.Errorf("unexpected hash result: want: %v got: %v", map[string][]byte{
			"ggg.test": expectedHash2,
			"gg.test":  expectedHash3,
		}, m2)
		return
	}
}

func TestLlcppg(t *testing.T) {
	os.Mkdir("testgenerate", 0777)
	defer os.RemoveAll("testgenerate")
	path, _ := filepath.Abs("testgenerate")
	generator := New(path, "cjson", path)

	os.WriteFile("testgenerate/llcppg.cfg", []byte(testLlcppgConfig), 0755)
	os.WriteFile("testgenerate/llpkg.cfg", []byte(testLLPkgConfig), 0755)

	cfg, err := config.ParseLLPkgConfig("testgenerate/llpkg.cfg")
	if err != nil {
		log.Fatalf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		log.Fatal(err)
	}
	_, err = uc.Installer.Install(uc.Pkg, "testgenerate")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("pkg-config", "--libs", "cjson")
	lockGoVersion(cmd, path)

	ret, _ := cmd.CombinedOutput()
	t.Log(string(ret))

	if err := generator.Generate(path); err != nil {
		t.Error(err)
		return
	}

	os.Mkdir(filepath.Join(path, ".generate"), 0777)

	if err := generator.Generate(filepath.Join(path, ".generate")); err != nil {
		t.Error(err)
		return
	}

	if err := generator.Check(filepath.Join(path, ".generate")); err != nil {
		t.Error(err)
		return
	}
	os.WriteFile("testgenerate/cJSON.go", []byte("1234"), 0755)
	if err := generator.Check(filepath.Join(path, ".generate")); err == nil {
		t.Error("unexpected check")
		return
	}
	// check go.mod
	checkGoMod(t, filepath.Join(path, ".generate", "go.mod"))
	checkGoMod(t, filepath.Join(path, "go.mod"))

	//generator.Check()
}
