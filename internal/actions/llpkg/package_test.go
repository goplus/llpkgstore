package llpkg

import (
	"os"
	"path/filepath"
	"testing"
)

func demoDir() (dir string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return
	}
	return filepath.Join(dir, "..", "..", "..", "_demo"), nil
	return
}

func checkName(t *testing.T, demoDir string, wantErr bool) {
	pkg, err := NewLLPkg(demoDir)
	if err != nil {
		if !wantErr {
			t.Error(err)
		}
		return
	}
	if wantErr {
		t.Error("unexpected no error")
		return
	}
	if pkg.Name() != "libcjson" {
		t.Errorf("unexpected package name: want %s got %s", "libcjson", pkg.Name())
	}

	if pkg.ClibName() != "cjson" {
		t.Errorf("unexpected llpkg package name: want %s got %s", "cjson", pkg.Name())
	}

	if pkg.ClibVersion() != "1.7.18" {
		t.Errorf("unexpected llpkg package version: want %s got %s", "1.7.18", pkg.ClibVersion())
	}
}

func TestReadConfig(t *testing.T) {
	demoDir, err := demoDir()
	if err != nil {
		t.Error(err)
		return
	}
	tempGoModFileName := filepath.Join(demoDir, "go.mod")

	t.Run("with-version-suffix", func(t *testing.T) {
		err := os.WriteFile(tempGoModFileName, []byte(`module github.com/goplus/llpkg/libcjson/v2

		go 1.22.0
		`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		defer os.Remove(tempGoModFileName)

		checkName(t, demoDir, false)
	})

	t.Run("without-version-suffix", func(t *testing.T) {
		err := os.WriteFile(tempGoModFileName, []byte(`module github.com/goplus/llpkg/libcjson

		go 1.22.0
		`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		defer os.Remove(tempGoModFileName)

		checkName(t, demoDir, false)
	})

	t.Run("raw-package-name", func(t *testing.T) {
		err := os.WriteFile(tempGoModFileName, []byte(`module libcjson

		go 1.22.0
		`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		defer os.Remove(tempGoModFileName)

		checkName(t, demoDir, false)

	})

	t.Run("wrong-go-mod", func(t *testing.T) {
		err := os.WriteFile(tempGoModFileName, []byte(`go 1.22.0`), 0644)
		if err != nil {
			t.Error(err)
			return
		}

		checkName(t, demoDir, true)
	})

	t.Run("no-go-mod", func(t *testing.T) {
		checkName(t, demoDir, true)
	})

}
