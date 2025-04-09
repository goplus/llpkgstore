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

func retrievePCNames(cppInfo map[string]cppInfo) (pcNames []string) {
	for name, info := range cppInfo {
		// skip itself
		if name == "root" {
			continue
		}
		if info.Properties.PkgName != "" {
			pcNames = append(pcNames, info.Properties.PkgName)
		}
	}
	return
}

func buildDependenciesMap(depsInfo dependencyInfo) map[string]string {
	deps := map[string]string{}

	// conan require info format: zlib/1.3.Z
	// 1.3.Z means match the major and minor version with 1.3
	for _, dep := range depsInfo.Requires {
		clib, version, ok := strings.Cut(dep, "/")
		if ok && version != "system" {
			deps[clib] = ""
		}
	}

	return deps
}

func resolveDependencies(graphInfo graphInfo) (dependencies []upstream.Package) {
	// 1. find require info from conan graph
	requireMap := buildDependenciesMap(graphInfo.Info)

	for _, info := range graphInfo.Dependencies {
		if info.Ref == "" {
			continue
		}
		clib, version, ok := strings.Cut(info.Ref, "/")
		if !ok {
			continue
		}
		// 2. fill the computed version from conan graph
		// by the way, remove duplicate elements.
		if _, ok := requireMap[clib]; ok {
			requireMap[clib] = version
		}
	}

	for clib, version := range requireMap {
		if version == "" {
			continue
		}
		dependencies = append(dependencies, upstream.Package{
			Name:    clib,
			Version: version,
		})
	}

	return
}

// in Conan, actual binary path is in the prefix field of *.pc file
func (c *conanInstaller) findBinaryPathFromPC(
	pkg upstream.Package,
	dir string,
	output []byte,
) (
	binaryDir string,
	pcName []string,
	err error,
) {
	var m installOutput
	err = json.Unmarshal(output, &m)
	if err != nil {
		return
	}

	if len(m.Graph.Nodes) == 0 {
		err = ErrPackageNotFound
		return
	}

	// default to package name,
	// first element is the real pkg-config name of this package
	// use append here to avoid resizing slice again.
	pcName = append(pcName, pkg.Name)

	for _, packageInfo := range m.Graph.Nodes {
		if packageInfo.Name == pkg.Name {
			// root must exist, this should not happen, returns an error.
			root, ok := packageInfo.CppInfo["root"]
			if !ok {
				err = ErrPackageNotFound
				return
			}
			if root.Properties.PkgName != "" {
				// root is the real pkg config name, replace instead.
				pcName[0] = root.Properties.PkgName
			}
			pcName = append(pcName, retrievePCNames(packageInfo.CppInfo)...)
			break
		}
	}

	pcFile, err := os.ReadFile(filepath.Join(dir, pcName[0]+".pc"))
	if err != nil {
		return
	}
	matches := pc.PrefixMatch.FindSubmatch(pcFile)
	if len(matches) != 2 {
		err = ErrPCFileNotFound
		return
	}
	binaryDir = string(matches[1])
	// check dir
	fs, err := os.Stat(binaryDir)
	if err != nil || !fs.IsDir() {
		if err == nil {
			err = ErrPCFileNotFound
		}
	}
	return
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
func (c *conanInstaller) Install(pkg upstream.Package, outputDir string) ([]string, error) {
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
		return nil, err
	}
	binaryDir, pkgConfigName, err := c.findBinaryPathFromPC(pkg, outputDir, ret)
	if err != nil {
		return nil, err
	}

	err = file.CopyFS(outputDir, os.DirFS(binaryDir), false)
	if err != nil {
		return nil, err
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

// Dependency retrieves the dependencies of a package using Conan's graph info command.
// It parses the dependency graph to extract required packages and their versions.
func (c *conanInstaller) Dependency(pkg upstream.Package) (dependencies []upstream.Package, err error) {
	// conan graph info --requires %s
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithConanSerializer())

	builder.SetName("conan")
	builder.SetSubcommand("graph")
	builder.SetObj("info")
	builder.SetArg("requires", pkg.Name+"/"+pkg.Version)
	builder.SetArg("format", "json")

	cmd := builder.Cmd()
	out, err := cmd.Output()
	if err != nil {
		return
	}

	var m graphOutput
	err = json.Unmarshal(out, &m)
	if err != nil {
		return
	}
	if len(m.Graph.Nodes) == 0 {
		err = ErrPackageNotFound
		return
	}

	for _, graphInfo := range m.Graph.Nodes {
		if graphInfo.Name == pkg.Name {
			dependencies = resolveDependencies(graphInfo)
			break
		}
	}
	return
}
