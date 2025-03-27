package githubrelease

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/goplus/llpkgstore/upstream"
)

func TestGHInstaller(t *testing.T) {
	ghr := &ghReleaseInstaller{
		config: map[string]string{
			"owner":    `MeteorsLiu`,
			"repo":     `llpkg`,
			"platform": `darwin`,
			"arch":     `amd64`,
		},
	}

	pkg := upstream.Package{
		Version: `v1.0.0`,
		Name:    `libxml2`,
	}

	tempDir, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		t.Errorf("Unexpected error when creating temp dir: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)

	if _, err = ghr.Install(pkg, "./f"); err != nil {
		t.Errorf("Install failed: %s", err)
	}

	if err := verify(pkg, tempDir); err != nil {
		t.Errorf("Verify failed: %s", err)
	}
}

func verify(pkg upstream.Package, installDir string) error {
	// 1. ensure .pc file exists
	_, err := os.Stat(filepath.Join(installDir,"lib/pkgconfig", pkg.Name+".pc"))
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
