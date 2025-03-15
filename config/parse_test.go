package config

import (
	"encoding/json"
	"testing"
)

const structJson = `{"upstream":{"installer":{"name":"conan"},"package":{"name":"cjson","version":"1.7.18"}}}`

func TestParseLLPkgConfig(t *testing.T) {
	config, err := ParseLLPkgConfig("../_demo/llpkg.cfg")
	if err != nil {
		t.Errorf("Error parsing config file: %v", err)
	}
	json, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Error marshaling config: %v", err)
	}
	if string(json) != structJson {
		t.Errorf("Unexpected config: %s", string(json))
	}
}
