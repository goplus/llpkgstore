package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseLLPkgConfig reads and parses the llpkg.cfg configuration file
// Performs the following operations:
//
// 1. Opens and reads the configuration file.
// 2. Deserializes JSON content into LLPkgConfig struct.
// 3. Applies default values for missing parameters.
// 4. Returns parsed config or I/O/decoding errors.
func ParseLLPkgConfig(configPath string) (LLPkgConfig, error) {
	var config LLPkgConfig
	file, err := os.Open(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to decode config file: %w", err)
	}

	// set default values
	config = fillDefaults(config)

	return config, nil
}

// fillDefaults applies default configuration values when parameters are missing.
// Current defaults:
// - installer.name: Uses first valid installer type if unspecified.
func fillDefaults(config LLPkgConfig) LLPkgConfig {
	if config.Upstream.Installer.Name == "" {
		config.Upstream.Installer.Name = ValidInstallers[0]
	}
	return config
}
