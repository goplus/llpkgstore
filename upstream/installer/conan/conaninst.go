package conan

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/goplus/llpkgstore/internal/cmdbuilder"
	"github.com/goplus/llpkgstore/upstream"
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

// conanInstaller implements the upstream.Installer interface using the Conan package manager.
// It handles installation of C/C++ libraries by executing installation commands,
// and managing dependencies through Conan's remote repositories.
type conanInstaller struct {
	config map[string]string
}

// NewConanInstaller creates a new Conan-based installer instance with provided configuration options.
// The config map supports custom Conan options (e.g., "options": "cjson:utils=True").
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

// options combines Conan default options with user-specified options from configuration
func (c *conanInstaller) options() string {
	return strings.Join([]string{"\\*:shared=True", c.config["options"]}, " ")
}

// Install executes Conan installation for the specified package into the output directory.
// It generates a conan install command with required options,
// and handles installation artifacts generation (e.g., .pc files).
func (c *conanInstaller) Install(pkg upstream.Package, outputDir string) error {
	// Build the following command
	// conan install --requires %s -g PkgConfigDeps --options \\*:shared=True --build=missing --output-folder=%s\
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithConanSerilazier())

	builder.SetName("conan")
	builder.SetSubcommand("install")
	builder.SetArg("requires", pkg.Name+"/"+pkg.Version)
	builder.SetArg("generator", "PkgConfigDeps")
	builder.SetArg("options", c.options())
	builder.SetArg("build", "missing")
	builder.SetArg("output-folder", outputDir)

	buildCmd := exec.Command(builder.Name(), append([]string{builder.Subcommand()}, builder.Args()...)...)

	out, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}

// Search checks Conan remote repository for the specified package availability.
// Returns the search results text and any encountered errors.
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
