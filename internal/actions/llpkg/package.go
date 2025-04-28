// Package llpkg provides utilities for managing language-linked packages (LLPkgs)
package llpkg

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/upstream"
)

var ErrWrongPackagePath = errors.New("llpkg: wrong package path")

func parsePackageName(goFile string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, goFile, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return file.Name.Name, nil
}

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

func FromPackageName(pkgName PackageName) (*LLPkg, error) {
	return NewLLPkg(pkgName.String())
}

// Returns the name of the package derived from the go.mod module path
func (p LLPkg) Name() PackageName {
	return PackageName(p.packageName)
}

// Returns the name from llpkg.cfg config
func (p LLPkg) ClibName() ClibName {
	return ClibName(p.cfg.Upstream.Package.Name)
}

// Returns the version from llpkg.cfg config
func (p LLPkg) ClibVersion() ClibVersion {
	return ClibVersion(p.cfg.Upstream.Package.Version)
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
	goFiles, _ := filepath.Glob(filepath.Join(p.packagePath, "*.go"))

	if len(goFiles) == 0 {
		err = ErrWrongPackagePath
		return
	}

	var packageName string

	for _, goFile := range goFiles {
		packageName, err = parsePackageName(goFile)
		if packageName != "" {
			break
		}
	}

	if packageName == "" {
		err = ErrWrongPackagePath
		return
	}

	p.cfg = cfg
	p.packageName = packageName
	return
}
