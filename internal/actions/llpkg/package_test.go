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
	testCases := []struct {
		name         string
		wantErr      bool
		goModContent []byte
	}{
		{
			name:    "with-version-suffix",
			wantErr: false,
			goModContent: []byte(`module github.com/goplus/llpkg/libcjson/v2

			go 1.22.0
			`),
		},
		{
			name:    "without-version-suffix",
			wantErr: false,
			goModContent: []byte(`module github.com/goplus/llpkg/libcjson

			go 1.22.0
			`),
		},
		{
			name:    "raw-package-name",
			wantErr: false,
			goModContent: []byte(`module libcjson

			go 1.22.0
			`),
		},
		{
			name:         "wrong-go-mod",
			wantErr:      true,
			goModContent: []byte(`go 1.22.0`),
		},
	}
	tempGoModFileName := filepath.Join(demoDir, "go.mod")

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := os.WriteFile(tempGoModFileName, tt.goModContent, 0644)
			if err != nil {
				t.Error(err)
				return
			}
			defer os.Remove(tempGoModFileName)

			checkName(t, demoDir, false)
		})
	}
	t.Run("no-go-mod", func(t *testing.T) {
		checkName(t, demoDir, true)
	})

}
