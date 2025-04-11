package conan

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"testing"

	"github.com/goplus/llpkgstore/internal/pc"
	"github.com/goplus/llpkgstore/upstream"
)

type packageSort []upstream.Package

func (vs packageSort) Len() int      { return len(vs) }
func (vs packageSort) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }

func (vs packageSort) Less(i, j int) bool {
	return vs[i].Version < vs[j].Version
}

func TestConanCJSON(t *testing.T) {
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

	sort.Strings(bp)
	if !reflect.DeepEqual(bp, []string{"cjson", "libcjson", "libcjson_utils"}) {
		t.Errorf("unexpected pc files: %v", bp)
		return
	}

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

	if !reflect.DeepEqual(bp, []string{"libxml-2.0"}) {
		t.Errorf("unexpected pc files: %v", bp)
		return
	}

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

func TestConanDependency(t *testing.T) {
	t.Run("fake", func(t *testing.T) {
		c := &conanInstaller{
			config: map[string]string{},
		}
		pkg := upstream.Package{
			Name:    "faketest1145141919",
			Version: "3.2.6",
		}
		_, err := c.Dependency(pkg)
		if err == nil {
			t.Errorf("unexpected behavior: no error")
		}
	})
	t.Run("sdl", func(t *testing.T) {
		c := &conanInstaller{
			config: map[string]string{},
		}
		pkg := upstream.Package{
			Name:    "sdl",
			Version: "3.2.6",
		}
		ver, err := c.Dependency(pkg)
		if err != nil {
			t.Error(err)
			return
		}
		sort.Sort(packageSort(ver))

		t.Log(ver)

		expectedDeps := []upstream.Package{
			{"libusb", "1.0.26"},
			{"dbus", "1.15.8"},
			{"libiconv", "1.17"},
			{"libalsa", "1.2.12"},
			{"wayland", "1.22.0"},
			{"zlib", "1.3.1"},
			{"xkbcommon", "1.6.0"},
			{"libsndio", "1.9.0"},
			{"pulseaudio", "17.0"},
			{"libxml2", "2.13.6"},
			{"expat", "2.7.1"},
			{"libffi", "3.4.4"},
		}
		if !reflect.DeepEqual(ver, expectedDeps) {
			t.Errorf("unexpected dependency for sdl: want %v got %v", expectedDeps, ver)
		}
	})

	t.Run("cjson", func(t *testing.T) {
		c := &conanInstaller{
			config: map[string]string{},
		}
		pkg := upstream.Package{
			Name:    "cjson",
			Version: "1.7.17",
		}
		ver, err := c.Dependency(pkg)
		if err != nil {
			t.Error(err)
			return
		}
		if len(ver) > 0 {
			t.Errorf("unexpected dependency for cjson: want %v got %v", []upstream.Package{}, ver)
		}
	})

	t.Run("libxml2", func(t *testing.T) {
		c := &conanInstaller{
			config: map[string]string{
				"options": `iconv=False`,
			},
		}
		pkg := upstream.Package{
			Name:    "libxml2",
			Version: "2.9.9",
		}
		ver, err := c.Dependency(pkg)
		if err != nil {
			t.Error(err)
			return
		}
		sort.Sort(packageSort(ver))

		expectedDeps := []upstream.Package{
			{"zlib", "1.3.1"},
		}
		if !reflect.DeepEqual(ver, expectedDeps) {
			t.Errorf("unexpected dependency for libxml2: want %v got %v", expectedDeps, ver)
		}
	})

	t.Run("libxslt-with-iconv", func(t *testing.T) {
		c := &conanInstaller{
			config: map[string]string{
				"options": `libxml2/*:iconv=False`,
			},
		}
		pkg := upstream.Package{
			Name:    "libxslt",
			Version: "1.1.42",
		}
		ver, err := c.Dependency(pkg)
		if err != nil {
			t.Error(err)
			return
		}
		sort.Sort(packageSort(ver))
		t.Log(ver)
		expectedDeps := []upstream.Package{
			{"zlib", "1.3.1"},
			{"libxml2", "2.13.6"},
		}
		if !reflect.DeepEqual(ver, expectedDeps) {
			t.Errorf("unexpected dependency for libxml2: want %v got %v", expectedDeps, ver)
		}
	})

	t.Run("libxslt-no-iconv", func(t *testing.T) {
		c := &conanInstaller{
			config: map[string]string{},
		}
		pkg := upstream.Package{
			Name:    "libxslt",
			Version: "1.1.42",
		}
		ver, err := c.Dependency(pkg)
		if err != nil {
			t.Error(err)
			return
		}
		sort.Sort(packageSort(ver))
		t.Log(ver)
		expectedDeps := []upstream.Package{
			{"libiconv", "1.17"},
			{"zlib", "1.3.1"},
			{"libxml2", "2.13.6"},
		}
		if !reflect.DeepEqual(ver, expectedDeps) {
			t.Errorf("unexpected dependency for libxml2: want %v got %v", expectedDeps, ver)
		}
	})
}

func verify(installDir string, pkgConfigName []string) error {
	for _, pkgName := range pkgConfigName {
		// 1. ensure .pc file exists
		_, err := os.Stat(filepath.Join(installDir, pkgName+".pc"))
		if err != nil {
			return errors.New(".pc file does not exist: " + err.Error())
		}
		absPath, err := filepath.Abs(installDir)
		if err != nil {
			return err
		}
		// 2. ensure pkg-config can find .pc file
		buildCmd := exec.Command("pkg-config", "--cflags", pkgName)

		pc.SetPath(buildCmd, absPath)
		out, err := buildCmd.CombinedOutput()
		if err != nil {
			return errors.New("pkg-config failed: " + err.Error() + " with output: " + string(out))
		}
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
