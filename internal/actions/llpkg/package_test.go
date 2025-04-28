package llpkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goplus/llpkgstore/internal/file"
)

func demoDir() (dir string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return
	}
	// ../../../_demo

	dir = filepath.Join("..", "..", "..", "_demo")
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

	t.Run("one-go-file", func(t *testing.T) {
		tempGoFileName := filepath.Join(demoDir, "x.go")
		err := os.WriteFile(tempGoFileName, []byte(`package libcjson`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		defer file.RemovePattern(filepath.Join(demoDir, "*.go"))

		checkName(t, demoDir, false)
	})

	t.Run("multi-go-files", func(t *testing.T) {
		tempGoFileName1 := filepath.Join(demoDir, "a.go")
		err := os.WriteFile(tempGoFileName1, []byte(`package libcjson`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		tempGoFileName2 := filepath.Join(demoDir, "x.go")
		err = os.WriteFile(tempGoFileName2, []byte(`package cjson`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		defer file.RemovePattern(filepath.Join(demoDir, "*.go"))

		checkName(t, demoDir, false)
	})

	t.Run("multi-go-files-fallback", func(t *testing.T) {
		tempGoFileName1 := filepath.Join(demoDir, "a.go")
		err := os.WriteFile(tempGoFileName1, []byte(``), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		tempGoFileName2 := filepath.Join(demoDir, "x.go")
		err = os.WriteFile(tempGoFileName2, []byte(`package libcjson`), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		defer file.RemovePattern(filepath.Join(demoDir, "*.go"))

		checkName(t, demoDir, false)
	})

	t.Run("no-go-files", func(t *testing.T) {
		checkName(t, demoDir, true)
	})

}
