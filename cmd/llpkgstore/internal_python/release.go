package internal

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/config"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build and upload Python binary packages",
	Long:  `Build and upload Python binary packages with simplified version handling`,
	RunE:  runPythonReleaseCmd,
}

func runPythonReleaseCmd(_ *cobra.Command, _ []string) error {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Check if llpkg.cfg exists and parse it
	cfgPath := filepath.Join(currentDir, "llpkg.cfg")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Errorf("llpkg.cfg not found in current directory")
	}

	// Parse the configuration to get Python package version
	cfg, err := config.ParseLLPkgConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to parse llpkg.cfg: %v", err)
	}

	// Direct version mapping: Python version -> Go version
	// e.g., numpy@1.26.4 -> v1.26.4
	pythonVersion := cfg.Upstream.Package.Version
	if pythonVersion == "" {
		return fmt.Errorf("package version not found in llpkg.cfg")
	}

	// Ensure version starts with 'v' for Go compatibility
	goVersion := pythonVersion
	if !strings.HasPrefix(goVersion, "v") {
		goVersion = "v" + goVersion
	}

	fmt.Printf("Starting Python package release with version: %s (mapped from Python version: %s)\n", goVersion, pythonVersion)

	// Check if generated files exist
	generatedFiles := []string{"go.mod", "go.sum"}
	for _, file := range generatedFiles {
		filePath := filepath.Join(currentDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file not found: %s", file)
		}
	}

	// Create artifact tar.gz with key files (*.go, go.mod, go.sum, llpyg.cfg, llpkg.cfg)
	base := filepath.Base(currentDir)
	artifactName := fmt.Sprintf("%s-%s.tar.gz", base, goVersion)
	artifactPath := filepath.Join(currentDir, artifactName)

	if err := createTarGz(artifactPath, currentDir, func(rel string) bool {
		// include root files only for simplicity
		// allow: *.go, go.mod, go.sum, llpyg.cfg, llpkg.cfg
		name := rel
		if strings.Contains(rel, string(filepath.Separator)) {
			// skip nested dirs to keep artifact small and predictable
			return false
		}
		if strings.HasSuffix(name, ".go") {
			return true
		}
		switch name {
		case "go.mod", "go.sum", "llpyg.cfg", "llpkg.cfg":
			return true
		}
		return false
	}); err != nil {
		return fmt.Errorf("failed to create artifact: %v", err)
	}

	// Export BIN_PATH and BIN_FILENAME to GITHUB_ENV for upload-artifact step
	if err := exportToGithubEnv("BIN_PATH", artifactPath); err != nil {
		return err
	}
	if err := exportToGithubEnv("BIN_FILENAME", artifactName); err != nil {
		return err
	}

	fmt.Printf("Python package release completed successfully with version: %s\n", goVersion)
	fmt.Printf("Note: This is a simplified release process for Python packages (%s@%s -> %s)\n", cfg.Upstream.Package.Name, pythonVersion, goVersion)

	return nil
}

// exportToGithubEnv writes key=value to $GITHUB_ENV so that subsequent steps can read it.
func exportToGithubEnv(key, value string) error {
	envFile := os.Getenv("GITHUB_ENV")
	if envFile == "" {
		// Local run fallback: print to stdout for debugging
		fmt.Printf("%s=%s\n", key, value)
		return nil
	}
	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_ENV: %v", err)
	}
	defer f.Close()
	if _, err := io.WriteString(f, fmt.Sprintf("%s=%s\n", key, value)); err != nil {
		return fmt.Errorf("failed to write to GITHUB_ENV: %v", err)
	}
	return nil
}

// createTarGz creates a tar.gz at dest, packing files from srcDir filtered by include(relPath).
func createTarGz(dest, srcDir string, include func(rel string) bool) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, ent := range entries {
		rel := ent.Name()
		if !include(rel) {
			continue
		}
		full := filepath.Join(srcDir, rel)
		info, err := os.Stat(full)
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}
		f, err := os.Open(full)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
