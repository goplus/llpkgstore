package config

import (
	"errors"

	"github.com/goplus/llpkgstore/upstream"
	"github.com/goplus/llpkgstore/upstream/installer/conan"
)

var ValidInstallers = []string{"conan"}

// LLPkgConfig represents the configuration structure parsed from llpkg.cfg files.
type LLPkgConfig struct {
	Upstream UpstreamConfig `json:"upstream"`
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
	default:
		return nil, errors.New("unknown upstream installer: " + upstreamConfig.Installer.Name)
	}
}
