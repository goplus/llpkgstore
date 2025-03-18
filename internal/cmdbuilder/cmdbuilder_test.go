package cmdbuilder

import (
	"testing"
)

func TestCmdBuilder(t *testing.T) {
	conanBuilder := NewCmdBuilder(WithConanSerializer())

	name := "conan"
	subcommand := "install"
	argMap := map[string]string{
		"requires": "cjson/1.7.18",
		"options":  `*:shared=True cjson/*:utils=True`,
		"build":    "missing",
	}
	args := []string{
		"--requires=cjson/1.7.18",
		`--options=*:shared=True cjson/*:utils=True`,
		"--build=missing",
	}

	conanBuilder.SetName(name)
	conanBuilder.SetSubcommand(subcommand)
	for k, v := range argMap {
		conanBuilder.SetArg(k, v)
	}

	if conanBuilder.Name() != name {
		t.Errorf("Unexpected name: %s", conanBuilder.Name())
	}

	if conanBuilder.Subcommand() != subcommand {
		t.Errorf("Unexpected subcommand: %s", conanBuilder.Subcommand())
	}

	for k, v := range args {
		if conanBuilder.args[k] != v {
			t.Errorf("Unexpected arg: %s", conanBuilder.args[k])
		}
	}
}
