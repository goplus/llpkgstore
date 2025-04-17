package actions

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/goplus/llpkgstore/config"
)

func TestBuildBinaryZip(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	// ../../_demo

	demoDir := filepath.Join(filepath.Dir(filepath.Dir(dir)), "_demo")

	cfg, err := config.ParseLLPkgConfig(filepath.Join(demoDir, "llpkg.cfg"))
	if err != nil {
		t.Error(err)
		return
	}

	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		t.Error(err)
		return
	}

	_, zipFilepath, err := BuildBinaryZip(uc)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(zipFilepath)

	zipr, err := zip.OpenReader(zipFilepath)
	if err != nil {
		t.Error(err)
		return
	}

	exceedMap := map[string]struct{}{
		"conaninfo.txt":                  {},
		"conanmanifest.txt":              {},
		"include/cjson/cJSON.h":          {},
		"lib/libcjson.so":                {},
		"lib/libcjson.so.1":              {},
		"lib/libcjson.so.1.7.18":         {},
		"lib/pkgconfig/cjson.pc.tmpl":    {},
		"lib/pkgconfig/libcjson.pc.tmpl": {},
		"licenses/LICENSE":               {},
	}

	if runtime.GOOS == "darwin" {
		exceedMap = map[string]struct{}{
			"conaninfo.txt":                                {},
			"conanmanifest.txt":                            {},
			"include/cjson/cJSON.h":                        {},
			"lib/cmake/conan-official-cjson-targets.cmake": {},
			"lib/libcjson.1.7.18.dylib":                    {},
			"lib/libcjson.1.dylib":                         {},
			"lib/libcjson.dylib":                           {},
			"lib/pkgconfig/cjson.pc.tmpl":                  {},
			"lib/pkgconfig/libcjson.pc.tmpl":               {},
			"licenses/LICENSE":                             {},
		}
	}

	fileMap := map[string]struct{}{}
	for _, file := range zipr.File {
		if !file.FileInfo().IsDir() {
			fileMap[file.Name] = struct{}{}
		}
	}

	for fileName := range exceedMap {
		if _, ok := fileMap[fileName]; !ok {
			t.Errorf("missing file: %s", fileName)
		}
	}

}
