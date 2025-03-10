package conan

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/goplus/llpkgstore/tools/internal/cmdbuilder"
	"github.com/goplus/llpkgstore/tools/pkg/upstream"
)

const (
	ConanfileTemplate = `[requires]
	%s/%s

	[generators]
	PkgConfigDeps

	[options]
	*:shared=True
	%s`
)

type conanInstaller struct {
	config map[string]string
}

func NewConanInstaller(config map[string]string) upstream.Installer {
	return &conanInstaller{
		config: config,
	}
}

func (c *conanInstaller) Name() string {
	return "conan"
}

func (c *conanInstaller) Config() map[string]string {
	return c.config
}

func (c *conanInstaller) options() string {
	return strings.Join([]string{"\\*:shared=True", c.config["options"]}, " ")
}

func (c *conanInstaller) Install(pkg upstream.Package, dir string) error {
	// Build the following command
	// conan install --requires %s -g PkgConfigDeps --options \\*:shared=True --build=missing --output-folder=%s\
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithConanSerilazier())

	builder.SetName("conan")
	builder.SetSubcommand("install")
	builder.SetArg("requires", pkg.Name+"/"+pkg.Version)
	builder.SetArg("generator", "PkgConfigDeps")
	builder.SetArg("options", c.options())
	builder.SetArg("build", "missing")
	builder.SetArg("output-folder", dir)

	buildCmd := exec.Command(builder.Name(), append([]string{builder.Subcommand()}, builder.Args()...)...)

	out, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}

func (c *conanInstaller) Search(pkg upstream.Package) (string, error) {
	// Build the following command
	// conan serach %s -r conancenter
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithConanSerilazier())

	builder.SetName("conan")
	builder.SetSubcommand("search")
	builder.SetObj(pkg.Name)
	builder.SetArg("remote", "conancenter")

	cmd := builder.Cmd()
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return "", err
	}

	return string(out), nil
}
