// Package llpkg provides utilities for managing language-linked packages (LLPkgs)
package llpkg

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/upstream"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

// Wraps an error with the "llpkg" prefix for better context
func wrapLLPkgError(err error) error {
	return fmt.Errorf("llpkg: %w", err)
}

// LLPkg represents a language-linked package with configuration and metadata
type LLPkg struct {
	cfg config.LLPkgConfig

	packageName string
	packagePath string
}

// Creates a new LLPkg instance by reading configuration from the specified path
func NewLLPkg(packagePath string) (*LLPkg, error) {
	if packagePath == "" {
		var err error
		packagePath, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	l := &LLPkg{packagePath: packagePath}

	if err := l.readPackageConfig(); err != nil {
		return nil, err
	}
	return l, nil
}

// Returns the name of the package derived from the go.mod module path
func (p LLPkg) Name() string {
	return p.packageName
}

// Returns the name from llpkg.cfg config
func (p LLPkg) ClibName() string {
	return p.cfg.Upstream.Package.Name
}

// Returns the version from llpkg.cfg config
func (p LLPkg) ClibVersion() string {
	return p.cfg.Upstream.Package.Version
}

// Retrieves the upstream source configuration for this package
func (p *LLPkg) Upstream() (*upstream.Upstream, error) {
	return config.NewUpstreamFromConfig(p.cfg.Upstream)
}

// Reads and validates the package configuration from llpkg.cfg and go.mod files
func (p *LLPkg) readPackageConfig() (err error) {
	llpkgFileName := filepath.Join(p.packagePath, "llpkg.cfg")
	cfg, err := config.ParseLLPkgConfig(llpkgFileName)
	if err != nil {
		err = wrapLLPkgError(err)
		return
	}
	goModFileName := filepath.Join(p.packagePath, "go.mod")

	goModContent, err := os.ReadFile(goModFileName)
	if err != nil {
		err = wrapLLPkgError(err)
		return
	}
	modFile, err := modfile.Parse(goModFileName, goModContent, nil)
	if err != nil {
		err = wrapLLPkgError(err)
		return
	}
	packageName := path.Base(modFile.Module.Mod.Path)

	if semver.IsValid(packageName) {
		// step forward if the last element is version
		// exmaple:
		// github.com/goplus/llpkg/cjson/v2
		// got v2
		// step forward: github.com/goplus/llpkg/cjson
		// got: cjson
		packageName = path.Base(path.Dir(modFile.Module.Mod.Path))
	}

	p.cfg = cfg
	p.packageName = packageName
	return
}
