package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goplus/llpkgstore/upstream"
	"github.com/goplus/llpkgstore/upstream/installer/conan"
	"github.com/goplus/llpkgstore/upstream/installer/pip"
)

var ValidInstallers = []string{"conan", "pip"}

// LLPkgConfig represents the configuration structure parsed from llpkg.cfg files.
type LLPkgConfig struct {
	Type     string         `json:"type,omitempty"` // "python" for Python packages, empty for C/C++
	Upstream UpstreamConfig `json:"upstream"`
	Llpyg    LlpygConfig    `json:"llpyg,omitempty"` // llpyg command line options
}

// LlpygConfig defines the llpyg command line options
type LlpygConfig struct {
	OutputDir string `json:"output_dir,omitempty"` // -o option, default "./test"
	ModName   string `json:"mod_name,omitempty"`   // -mod option, default package name
	ModDepth  int    `json:"mod_depth,omitempty"`  // -d option, default 1
}

// Validate validates the LlpygConfig
func (l *LlpygConfig) Validate() error {
	if l.ModDepth < 0 {
		return errors.New("mod_depth must be non-negative")
	}
	if l.ModDepth > 10 {
		return errors.New("mod_depth should not exceed 10 for performance reasons")
	}

	// Validate output directory path
	if l.OutputDir != "" {
		if err := validateOutputDir(l.OutputDir); err != nil {
			return fmt.Errorf("invalid output_dir: %v", err)
		}
	}

	// Validate module name
	if l.ModName != "" {
		if err := validateModuleName(l.ModName); err != nil {
			return fmt.Errorf("invalid mod_name: %v", err)
		}
	}

	return nil
}

// validateOutputDir validates the output directory path
func validateOutputDir(outputDir string) error {
	// Check if path contains illegal characters
	illegalChars := []string{"..", "~", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range illegalChars {
		if strings.Contains(outputDir, char) {
			return fmt.Errorf("output directory contains illegal character: %s", char)
		}
	}

	// Check path length
	if len(outputDir) > 255 {
		return errors.New("output directory path too long")
	}

	return nil
}

// validateModuleName validates the Go module name
func validateModuleName(modName string) error {
	// Check module name format
	if !strings.Contains(modName, "/") {
		return errors.New("module name should contain at least one slash")
	}

	// Check if it starts with github.com or other valid domains
	validPrefixes := []string{"github.com", "gitlab.com", "gitee.com", "bitbucket.org"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(modName, prefix+"/") {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("module name should start with a valid domain (e.g., github.com)")
	}

	// Check module name length
	if len(modName) > 200 {
		return errors.New("module name too long")
	}

	return nil
}

// GetDefaultModDepth returns the default module depth if not specified
func (l *LlpygConfig) GetDefaultModDepth() int {
	if l.ModDepth == 0 {
		return 1 // Default depth is 1
	}
	return l.ModDepth
}

// GetDefaultOutputDir returns the default output directory if not specified
func (l *LlpygConfig) GetDefaultOutputDir() string {
	if l.OutputDir == "" {
		return "./test" // Default output directory
	}
	return l.OutputDir
}

// GetDefaultModName returns the default module name if not specified
func (l *LlpygConfig) GetDefaultModName() string {
	if l.ModName == "" {
		return "" // Default to use package name
	}
	return l.ModName
}

// UpstreamConfig defines the upstream configuration containing installer settings and package metadata.
type UpstreamConfig struct {
	Installer InstallerConfig `json:"installer"`
	Package   PackageConfig   `json:"package"`
}

// InstallerConfig specifies the installer type and its configuration options.
// "name" field must match supported installers (e.g., "conan").
// "config" holds installer-specific parameters (optional).
type InstallerConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config,omitempty"`
}

// PackageConfig defines the target library package's identifier and version requirements.
type PackageConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewUpstreamFromConfig creates an Upstream instance from configuration data.
// Returns error if unsupported installer type is specified.
func NewUpstreamFromConfig(upstreamConfig UpstreamConfig) (*upstream.Upstream, error) {
	switch upstreamConfig.Installer.Name {
	case "conan":
		return &upstream.Upstream{
			Installer: conan.NewConanInstaller(upstreamConfig.Installer.Config),
			Pkg: upstream.Package{
				Name:    upstreamConfig.Package.Name,
				Version: upstreamConfig.Package.Version,
			},
		}, nil
	case "pip":
		return &upstream.Upstream{
			Installer: pip.NewPipInstaller(upstreamConfig.Installer.Config),
			Pkg: upstream.Package{
				Name:    upstreamConfig.Package.Name,
				Version: upstreamConfig.Package.Version,
			},
		}, nil
	default:
		return nil, errors.New("unknown upstream installer: " + upstreamConfig.Installer.Name)
	}
}
