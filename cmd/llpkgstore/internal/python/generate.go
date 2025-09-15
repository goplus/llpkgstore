package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/generator/llpyg"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Python bindings",
	Long:  ``,
	RunE:  runLLPygGenerate,
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// isPackageInstalledInSystem checks if the specified package is already installed in the system environment
func isPackageInstalledInSystem(packageName string) bool {
	// Method 1: Try to import the package directly
	if canImportPackage(packageName) {
		return true
	}

	// Method 2: Check pip list output
	if isPackageInPipList(packageName) {
		return true
	}

	return false
}

// canImportPackage tries to import the package to check if it's installed
func canImportPackage(packageName string) bool {
	cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s; print('OK')", packageName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Package %s import test failed: %v", packageName, err)
		return false
	}

	// Check if output contains "OK"
	result := strings.TrimSpace(string(output))
	return strings.Contains(result, "OK")
}

// isPackageInPipList checks if the package is in pip list
func isPackageInPipList(packageName string) bool {
	cmd := exec.Command("pip3", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to run pip3 list: %v", err)
		return false
	}

	// Check if package name is in the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, packageName) {
			return true
		}
	}

	return false
}

func runLLPygGenerateWithDir(dir string) error {
	cfg, err := config.ParseLLPkgConfig(filepath.Join(dir, LLGOModuleIdentifyFile))
	if err != nil {
		return fmt.Errorf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		return err
	}
	log.Printf("Start to generate %s", uc.Pkg.Name)

	// Prioritize checking if package is already installed in system environment
	var pythonDir string
	var tempDir string
	var needCleanup bool

	// Check if the package already exists in system environment
	if isPackageInstalledInSystem(uc.Pkg.Name) {
		log.Printf("Package %s found in system environment, using system installation", uc.Pkg.Name)
		pythonDir = "" // Use system environment, no need to set PYTHONPATH
	} else {
		log.Printf("Package %s not found in system environment, installing to temporary directory", uc.Pkg.Name)
		tempDir, err = os.MkdirTemp("", "llpkg-tool")
		if err != nil {
			return err
		}
		needCleanup = true
		defer func() {
			if needCleanup {
				os.RemoveAll(tempDir)
			}
		}()

		_, err = uc.Installer.Install(uc.Pkg, tempDir)
		if err != nil {
			return err
		}
		pythonDir = tempDir
	}

	// Check if this is a Python package
	if cfg.Type == "python" {
		// For Python packages, directly use llpyg generator
		// This will call "llpyg numpy" and copy the generated files
		generator := llpyg.New(dir, cfg.Upstream.Package.Name, pythonDir)
		return generator.Generate(dir)
	} else {
		// For C/C++ packages, we need to import the C++ generator
		// This is a simplified version - in practice you might want to handle this differently
		return fmt.Errorf("C/C++ packages not supported in Python version")
	}
}

func runLLPygGenerate(_ *cobra.Command, args []string) error {
	// Detect environment based on the first directory
	path := currentDir()
	if len(args) > 0 {
		if absPath, err := filepath.Abs(args[0]); err == nil {
			path = absPath
		}
	}

	// Check if this is a Python package by reading the config
	cfg, err := config.ParseLLPkgConfig(filepath.Join(path, LLGOModuleIdentifyFile))
	if err == nil && cfg.Type == "python" {
		// For Python packages, we don't need conan profile detection
		log.Printf("Detected Python package: %s", cfg.Upstream.Package.Name)
	} else {
		// For C/C++ packages, detect conan profile
		exec.Command("conan", "profile", "detect").Run()
	}

	// by default, use current dir
	if len(args) == 0 {
		return runLLPygGenerateWithDir(path)
	}
	for _, argPath := range args {
		absPath, err := filepath.Abs(argPath)
		if err != nil {
			continue
		}
		err = runLLPygGenerateWithDir(absPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
