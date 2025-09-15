package internal

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build and upload Python binary packages",
	Long:  `Build and upload Python binary packages with unified version management`,
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

	// Use DefaultClient for unified version management (same as C++)
	client, err := actions.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %v", err)
	}

	// Use the same release logic as C++ (unified version management)
	fmt.Println("Using unified release logic (same as C++)...")
	if err := client.Release(); err != nil {
		return fmt.Errorf("failed to run release: %v", err)
	}

	pythonVersion := cfg.Upstream.Package.Version
	packageName := cfg.Upstream.Package.Name

	fmt.Printf("Python package release completed successfully\n")
	fmt.Printf("Package: %s, Python Version: %s\n", packageName, pythonVersion)
	fmt.Printf("Note: Now using unified version management (same as C++)")

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
