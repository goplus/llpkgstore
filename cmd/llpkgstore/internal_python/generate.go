package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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

	tempDir, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	_, err = uc.Installer.Install(uc.Pkg, tempDir)
	if err != nil {
		return err
	}

	// Check if this is a Python package
	if cfg.Type == "python" {
		// For Python packages, directly use llpyg generator
		// This will call "llpyg numpy" and copy the generated files
		generator := llpyg.New(dir, cfg.Upstream.Package.Name, tempDir)
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
