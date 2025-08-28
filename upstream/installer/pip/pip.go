package pip

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goplus/llpkgstore/internal/cmdbuilder"
	"github.com/goplus/llpkgstore/upstream"
)

var (
	ErrPackageNotFound = errors.New("package not found")
	ErrPythonNotFound  = errors.New("python not found")
)

// pipInstaller implements the upstream.Installer interface using pip package manager.
// It handles installation of Python libraries by executing pip install commands.
type pipInstaller struct {
	config map[string]string
}

// NewPipInstaller creates a new pip-based installer instance with provided configuration options.
func NewPipInstaller(config map[string]string) upstream.Installer {
	return &pipInstaller{
		config: config,
	}
}

func (p *pipInstaller) Name() string {
	return "pip"
}

func (p *pipInstaller) Config() map[string]string {
	return p.config
}

// options combines pip default options with user-specified options from configuration
func (p *pipInstaller) options() []string {
	return strings.Fields(p.config["options"])
}

// Install executes pip installation for the specified package into the output directory.
// It generates a pip install command with required options.
func (p *pipInstaller) Install(pkg upstream.Package, outputDir string) ([]string, error) {
	// Build the following command
	// pip3 install --target=%s %s==%s
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithPipSerializer())

	builder.SetName("pip3")
	builder.SetSubcommand("install")
	builder.SetArg("target", outputDir)
	builder.SetObj(pkg.Name + "==" + pkg.Version)

	for _, opt := range p.options() {
		builder.SetArg("options", opt)
	}

	buildCmd := builder.Cmd()
	buildCmd.Stderr = os.Stderr
	ret, err := buildCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pip install failed: %v, output: %s", err, string(ret))
	}

	// For Python packages, we return the package name as the "config file"
	// since Python doesn't use pkg-config files like C/C++
	return []string{pkg.Name}, nil
}

// Search checks pip repository for the specified package availability.
// Returns the search results text and any encountered errors.
func (p *pipInstaller) Search(pkg upstream.Package) ([]string, error) {
	// Build the following command
	// pip3 search %s
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithPipSerializer())

	builder.SetName("pip3")
	builder.SetSubcommand("search")
	builder.SetObj(pkg.Name)

	cmd := builder.Cmd()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pip search failed: %v", err)
	}

	if strings.Contains(string(out), "not found") {
		return nil, ErrPackageNotFound
	}

	var ret []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, pkg.Name) {
			ret = append(ret, strings.TrimSpace(line))
		}
	}

	return ret, nil
}

// Dependency retrieves the dependencies of a package using pip show command.
// It parses the package information to extract required packages and their versions.
func (p *pipInstaller) Dependency(pkg upstream.Package) (dependencies []upstream.Package, err error) {
	// pip3 show %s
	builder := cmdbuilder.NewCmdBuilder(cmdbuilder.WithPipSerializer())

	builder.SetName("pip3")
	builder.SetSubcommand("show")
	builder.SetObj(pkg.Name)

	var pipError bytes.Buffer

	cmd := builder.Cmd()
	cmd.Stderr = &pipError

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pip show failed: %v, error: %s", err, pipError.String())
	}

	// Parse pip show output to extract dependencies
	// This is a simplified implementation - in practice, you might want to use
	// pip list --format=json or similar for better parsing
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Requires:") {
			requires := strings.TrimSpace(strings.TrimPrefix(line, "Requires:"))
			if requires != "" {
				deps := strings.Split(requires, ",")
				for _, dep := range deps {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						// Parse dependency name and version
						parts := strings.Split(dep, " ")
						if len(parts) >= 1 {
							dependencies = append(dependencies, upstream.Package{
								Name:    parts[0],
								Version: "", // Version info might be in a different format
							})
						}
					}
				}
			}
			break
		}
	}

	return dependencies, nil
}
