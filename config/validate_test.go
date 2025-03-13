package config

import "testing"

func TestValidateLLPkgConfig(t *testing.T) {
	config, err := ParseLLPkgConfig("../_demo/llpkg.cfg")
	if err != nil {
		t.Errorf("Error parsing config file: %v", err)
	}
	err = ValidateLLPkgConfig(config)
	if err != nil {
		t.Errorf("Error validating config: %v", err)
	}
}
