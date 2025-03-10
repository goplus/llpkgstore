package config

import (
	"errors"

	"github.com/goplus/llpkgstore/tools/pkg/upstream"
	"github.com/goplus/llpkgstore/tools/pkg/upstream/installer/conan"
)

var ValidInstallers = []string{"conan"}

type LLpkgConfig struct {
	UpstreamConfig UpstreamConfig `json:"upstream"`
}

type UpstreamConfig struct {
	InstallerConfig InstallerConfig `json:"installer"`
	PackageConfig   PackageConfig   `json:"package"`
}

type InstallerConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config,omitempty"`
}

type PackageConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewUpstreamFromConfig(upstreamConfig UpstreamConfig) (*upstream.Upstream, error) {
	switch upstreamConfig.InstallerConfig.Name {
	case "conan":
		return upstream.NewUpstream(conan.NewConanInstaller(upstreamConfig.InstallerConfig.Config), upstream.Package{
			Name:    upstreamConfig.PackageConfig.Name,
			Version: upstreamConfig.PackageConfig.Version,
		}), nil
	default:
		return nil, errors.New("unknown upstream installer: " + upstreamConfig.InstallerConfig.Name)
	}
}
