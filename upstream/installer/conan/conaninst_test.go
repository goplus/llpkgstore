package conan

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

	if err := c.Install(pkg, tempDir); err != nil {
		t.Errorf("Install failed: %s", err)
	}

	if err := verify(pkg, tempDir); err != nil {
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
}

func verify(pkg upstream.Package, installDir string) error {
	// 1. ensure .pc file exists
	_, err := os.Stat(filepath.Join(installDir, pkg.Name+".pc"))
	if err != nil {
		return errors.New(".pc file does not exist: " + err.Error())
	}

	// 2. ensure pkg-config can find .pc file
	os.Setenv("PKG_CONFIG_PATH", installDir)
	defer os.Unsetenv("PKG_CONFIG_PATH")

	buildCmd := exec.Command("pkg-config", "--cflags", pkg.Name)
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return errors.New("pkg-config failed: " + err.Error() + " with output: " + string(out))
	}

	return nil
}
