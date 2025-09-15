package pip

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	fmt.Printf("Installing Python package: %s==%s to %s\n", pkg.Name, pkg.Version, outputDir)

	// Check if output directory exists, create if not
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %v", outputDir, err)
	}

	// Build the following command
	// pip3 install --target=%s --no-deps --no-cache-dir %s==%s
	args := []string{"install", "--target", outputDir}

	// Add additional pip options
	for _, opt := range p.options() {
		args = append(args, opt)
	}

	// Add some common pip options for better stability
	args = append(args, "--no-deps")      // Temporarily don't install dependencies to avoid version conflicts
	args = append(args, "--no-cache-dir") // Don't use cache to ensure getting the latest version

	// Add package name and version
	args = append(args, pkg.Name+"=="+pkg.Version)

	buildCmd := exec.Command("pip3", args...)
	buildCmd.Stderr = os.Stderr

	fmt.Printf("Executing pip install command...\n")
	ret, err := buildCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pip install failed for package %s==%s: %v, output: %s",
			pkg.Name, pkg.Version, err, string(ret))
	}

	fmt.Printf("Successfully installed Python package: %s==%s\n", pkg.Name, pkg.Version)

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
	fmt.Printf("Retrieving dependencies for package: %s\n", pkg.Name)

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
		return nil, fmt.Errorf("pip show failed for package %s: %v, error: %s",
			pkg.Name, err, pipError.String())
	}

	// Parse pip show output to extract dependencies
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Requires:") {
			requires := strings.TrimSpace(strings.TrimPrefix(line, "Requires:"))
			if requires != "" && requires != "None" {
				deps := strings.Split(requires, ",")
				for _, dep := range deps {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						// Parse dependency name and version
						// Handle formats like "numpy>=1.20.0", "requests==2.31.0", "pandas"
						var depName, depVersion string

						// Check for version specifiers
						if strings.Contains(dep, ">=") {
							parts := strings.Split(dep, ">=")
							depName = strings.TrimSpace(parts[0])
							depVersion = strings.TrimSpace(parts[1])
						} else if strings.Contains(dep, "==") {
							parts := strings.Split(dep, "==")
							depName = strings.TrimSpace(parts[0])
							depVersion = strings.TrimSpace(parts[1])
						} else if strings.Contains(dep, ">") {
							parts := strings.Split(dep, ">")
							depName = strings.TrimSpace(parts[0])
							depVersion = strings.TrimSpace(parts[1])
						} else if strings.Contains(dep, "<=") {
							parts := strings.Split(dep, "<=")
							depName = strings.TrimSpace(parts[0])
							depVersion = strings.TrimSpace(parts[1])
						} else if strings.Contains(dep, "<") {
							parts := strings.Split(dep, "<")
							depName = strings.TrimSpace(parts[0])
							depVersion = strings.TrimSpace(parts[1])
						} else {
							// No version specifier, just package name
							depName = dep
							depVersion = ""
						}

						if depName != "" {
							dependencies = append(dependencies, upstream.Package{
								Name:    depName,
								Version: depVersion,
							})
							fmt.Printf("Found dependency: %s (version: %s)\n", depName, depVersion)
						}
					}
				}
			}
			break
		}
	}

	fmt.Printf("Total dependencies found: %d\n", len(dependencies))
	return dependencies, nil
}
