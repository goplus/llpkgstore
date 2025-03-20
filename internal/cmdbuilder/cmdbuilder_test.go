package cmdbuilder

import (
	"slices"
	"testing"
)

func TestCmdBuilder(t *testing.T) {
	conanBuilder := NewCmdBuilder(WithConanSerializer())

	// Provided
	name := "conan"
	subcommand := "install"
	inputArgMap := map[string]string{
		"requires": "cjson/1.7.18",
		"options":  `*:shared=True cjson/*:utils=True`,
		"build":    "missing",
	}

	// Expected
	expectedArgs := []string{
		"--requires=cjson/1.7.18",
		`--options=*:shared=True cjson/*:utils=True`,
		"--build=missing",
	}

	conanBuilder.SetName(name)
	conanBuilder.SetSubcommand(subcommand)
	for name, value := range inputArgMap {
		conanBuilder.SetArg(name, value)
	}

	if conanBuilder.Name() != name {
		t.Errorf("Unexpected name: %s", conanBuilder.Name())
	}

	if conanBuilder.Subcommand() != subcommand {
		t.Errorf("Unexpected subcommand: %s", conanBuilder.Subcommand())
	}

	for _, arg := range expectedArgs {
		if !slices.Contains(conanBuilder.Args(), arg) {
			t.Errorf("No expected arg: %s", arg)
		}
	}
}
