package llcppg

import (
	"encoding/hex"
	"os"
	"reflect"
	"testing"
)

const (
	testLlpkgConfig = `{
  "upstream": {
    "package": {
      "name": "cjson",
      "version": "1.7.18"
    }
  }
}`
	testLlcppgConfig = `{
		"name": "libcjson",
		"cflags": "$(pkg-config --cflags libcjson)",
		"libs": "$(pkg-config --libs libcjson)",
		"include": [
			"cjson/cJSON.h"
		],
		"deps": null,
		"trimPrefixes": [],
		"cplusplus": false
	}`
)

func TestFindGoMod(t *testing.T) {
	goModFile = "ggg.test"
	l := &llcppgGenerator{dir: "."}
	if ret := l.findGoMod(); ret != "testfind2/testfind" {
		t.Errorf("unexpected find result: got %s want: testfind", ret)
	}
}

func TestHash(t *testing.T) {
	canHashFile["gg.test"] = struct{}{}
	canHashFile["ggg.test"] = struct{}{}
	m, err := hashDir("testfind2")
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
	m2, err := hashDir("testfind2/testfind")
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
	goModFile = "go.mod"
	os.Mkdir("testgenerate", 0777)
	defer os.RemoveAll("testgenerate")
	generator := New("testgenerate")
	os.WriteFile("testgenerate/llcppg.cfg", []byte(testLlcppgConfig), 0755)
	os.WriteFile("testgenerate/llpkg.cfg", []byte(testLlpkgConfig), 0755)

	if err := generator.Generate(); err != nil {
		t.Error(err)
		return
	}
	// copy out
	os.CopyFS("testgenerate", os.DirFS("testgenerate/libcjson"))

	baseDir := generator.(*llcppgGenerator).findGoMod()
	if baseDir != "testgenerate/libcjson" {
		t.Errorf("unexpected generate dir: want: testgenerate/libcjson got: %s", baseDir)
		return
	}
	if err := generator.Check(); err != nil {
		t.Error(err)
		return
	}
	os.WriteFile("testgenerate/cJSON.go", []byte("1234"), 0755)
	if err := generator.Check(); err == nil {
		t.Error("unexpected check")
		return
	}
	//generator.Check()
}
