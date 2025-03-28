package conan

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/internal/cmdbuilder"
	"github.com/goplus/llpkgstore/internal/file"
	"github.com/goplus/llpkgstore/internal/pc"
	"github.com/goplus/llpkgstore/upstream"
)

var (
	ErrPackageNotFound = errors.New("package not found")
	ErrPCFileNotFound  = errors.New("pc file not found")
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

// in Conan, actual binary path is in the prefix field of *.pc file
func (c *conanInstaller) findBinaryPathFromPC(pkg upstream.Package, dir string, installOutput []byte) (string, string, error) {
	var m conanOutput
	json.Unmarshal(installOutput, &m)

	if len(m.Graph.Nodes) == 0 {
		return "", "", ErrPackageNotFound
	}
	// default to package name.
	pkgConfigName := pkg.Name

	for _, packageInfo := range m.Graph.Nodes {
		realPCName := packageInfo.CppInfo.Root.Properties.PkgName

		if packageInfo.Name == pkg.Name && realPCName != "" {
			// ok this is the result we want
			pkgConfigName = realPCName
			break
		}
	}

	pcFile, err := os.ReadFile(filepath.Join(dir, pkgConfigName+".pc"))
	if err != nil {
		return "", "", err
	}
	matches := pc.PrefixMatch.FindSubmatch(pcFile)
	if len(matches) != 2 {
		return "", "", ErrPCFileNotFound
	}
	binaryDir := string(matches[1])
	// check dir
	fs, err := os.Stat(binaryDir)
	if err != nil || !fs.IsDir() {
		return "", "", ErrPCFileNotFound
	}
	return binaryDir, pkgConfigName, nil
}

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
func (c *conanInstaller) options() []string {
	arr := strings.Join([]string{`*:shared=True`, c.config["options"]}, " ")
	return strings.Fields(arr)
}

// Install executes Conan installation for the specified package into the output directory.
// It generates a conan install command with required options,
// and handles installation artifacts generation (e.g., .pc files).
func (c *conanInstaller) Install(pkg upstream.Package, outputDir string) (string, error) {
	// Build the following command
	// conan install --requires %s -g PkgConfigDeps --options \\*:shared=True --build=missing --output-folder=%s\
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithConanSerializer())

	builder.SetName("conan")
	builder.SetSubcommand("install")
	builder.SetArg("requires", pkg.Name+"/"+pkg.Version)
	builder.SetArg("generator", "PkgConfigDeps")
	builder.SetArg("build", "missing")
	builder.SetArg("output-folder", outputDir)
	builder.SetArg("format", "json")

	for _, opt := range c.options() {
		builder.SetArg("options", opt)
	}

	buildCmd := builder.Cmd()

	// conan will output install result to Stdout, output progress to Stderr
	buildCmd.Stderr = os.Stderr
	ret, err := buildCmd.Output()
	if err != nil {
		// fmt.Println(string(out))
		return "", err
	}
	binaryDir, pkgConfigName, err := c.findBinaryPathFromPC(pkg, outputDir, ret)
	if err != nil {
		return "", err
	}

	err = file.CopyFS(outputDir, os.DirFS(binaryDir), false)
	if err != nil {
		return "", err
	}

	return pkgConfigName, nil
}

// Search checks Conan remote repository for the specified package availability.
// Returns the search results text and any encountered errors.
func (c *conanInstaller) Search(pkg upstream.Package) ([]string, error) {
	// Build the following command
	// conan search %s -r conancenter
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithConanSerializer())

	builder.SetName("conan")
	builder.SetSubcommand("search")
	builder.SetObj(pkg.Name)
	builder.SetArg("remote", "conancenter")

	cmd := builder.Cmd()
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return nil, err
	}
	if strings.Contains(string(out), "not found") {
		return nil, ErrPackageNotFound
	}

	var ret []string

	for _, field := range strings.Fields(string(out)) {
		prefix, _, found := strings.Cut(field, "/")
		if found && prefix == pkg.Name {
			ret = append(ret, field)
		}
	}

	return ret, nil
}
