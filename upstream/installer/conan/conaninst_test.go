package conan

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/goplus/llpkgstore/upstream"
)

func TestConanInstaller(t *testing.T) {
	c := &conanInstaller{
		config: map[string]string{
			"options": `cjson/*:utils=True`,
		},
	}

	pkg := upstream.Package{
		Name:    "cjson",
		Version: "1.7.18",
	}

	if name := c.Name(); name != "conan" {
		t.Errorf("Unexpected name: %s", name)
	}

	tempDir, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		t.Errorf("Unexpected error when creating temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	bp, err := c.Install(pkg, tempDir)
	if err != nil {
		t.Errorf("Install failed: %s", err)
	}

	t.Log(bp)

	if err := verify(tempDir, bp); err != nil {
		t.Errorf("Verify failed: %s", err)
	}
}

// https://github.com/goplus/llpkgstore/issues/19
func TestConanIssue19(t *testing.T) {
	c := &conanInstaller{
		config: map[string]string{},
	}

	pkg := upstream.Package{
		Name:    "libxml2",
		Version: "2.9.9",
	}

	if name := c.Name(); name != "conan" {
		t.Errorf("Unexpected name: %s", name)
	}

	tempDir, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		t.Errorf("Unexpected error when creating temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	bp, err := c.Install(pkg, tempDir)
	if err != nil {
		t.Errorf("Install failed: %s", err)
	}

	t.Log(bp)

	if err := verify(tempDir, bp); err != nil {
		t.Errorf("Verify failed: %s", err)
	}
}

func TestConanSearch(t *testing.T) {
	c := &conanInstaller{
		config: map[string]string{
			"options": `cjson/*:utils=True`,
		},
	}

	pkg := upstream.Package{
		Name:    "cjson",
		Version: "1.7.18",
	}
	ver, _ := c.Search(pkg)
	if !slices.Contains(ver, "cjson/1.7.18") {
		t.Errorf("unexpected search result: %s", ver)
	}

	t.Log(ver)

	pkg = upstream.Package{
		Name:    "cjson2",
		Version: "1.7.18",
	}

	_, err := c.Search(pkg)
	if err == nil {
		t.Errorf("unexpected behavior: %s", err)
	}

}

func verify(installDir, pkgConfigName string) error {
	// 1. ensure .pc file exists
	_, err := os.Stat(filepath.Join(installDir, pkgConfigName+".pc"))
	if err != nil {
		return errors.New(".pc file does not exist: " + err.Error())
	}

	// 2. ensure pkg-config can find .pc file
	os.Setenv("PKG_CONFIG_PATH", installDir)
	defer os.Unsetenv("PKG_CONFIG_PATH")

	buildCmd := exec.Command("pkg-config", "--cflags", pkgConfigName)
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return errors.New("pkg-config failed: " + err.Error() + " with output: " + string(out))
	}

	switch runtime.GOOS {
	case "linux":
		matches, _ := filepath.Glob(filepath.Join(installDir, "lib", "*.so"))
		if len(matches) == 0 {
			return errors.New("cannot find so file")
		}
	case "darwin":
		matches, _ := filepath.Glob(filepath.Join(installDir, "lib", "*.dylib"))
		if len(matches) == 0 {
			return errors.New("cannot find dylib file")
		}
	}

	return nil
}
