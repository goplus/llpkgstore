package config

import (
	"fmt"
	"slices"
)

// ValidateLLPkgConfig performs structural validation of the configuration.
// Validates upstream installer and package metadata requirements.
func ValidateLLPkgConfig(config LLPkgConfig) error {
	return validateUpstreamConfig(config.Upstream)
}

// validateUpstreamConfig performs detailed validation of upstream configuration parameters.
func validateUpstreamConfig(config UpstreamConfig) error {
	// 1. check if upstream installer is valid
	if config.Installer.Name == "" {
		return fmt.Errorf("missing required installer type: upstream.installer.name must be specified")
	}
	if !slices.Contains(ValidInstallers, config.Installer.Name) {
		return fmt.Errorf("unsupported installer type: %s (valid options: %v)", config.Installer.Name, ValidInstallers)
	}

	// 2. check if package is valid
	if config.Package.Name == "" {
		return fmt.Errorf("missing required package identifier: upstream.package.name cannot be empty")
	}
	if config.Package.Version == "" {
		return fmt.Errorf("missing required version specification: upstream.package.version cannot be empty")
	}

	return nil
}
