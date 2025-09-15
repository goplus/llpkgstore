package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var postProcessingCmd = &cobra.Command{
	Use:   "postprocessing",
	Short: "Process merged PR for Python packages",
	Long:  `Process merged PR for Python packages with unified version management`,
	RunE:  runPythonPostProcessingCmd,
}

// LLPkgStoreJSON represents the structure of llpkgstore.json (C++ compatible)
type LLPkgStoreJSON struct {
	Packages map[string]PackageInfo `json:"packages"`
}

// PackageInfo represents package version information (C++ compatible)
type PackageInfo struct {
	Versions []VersionInfo `json:"versions"`
}

// VersionInfo represents version mapping (C++ compatible)
type VersionInfo struct {
	Python string   `json:"python"`
	Go     []string `json:"go"`
}

func runPythonPostProcessingCmd(_ *cobra.Command, _ []string) error {
	fmt.Println("Starting Python package post-processing with unified version management...")

	// Use DefaultClient for unified version management (same as C++)
	client, err := actions.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %v", err)
	}

	// Use the same postprocessing logic as C++ (unified version management)
	fmt.Println("Using unified postprocessing logic (same as C++)...")
	if err := client.Postprocessing(); err != nil {
		return fmt.Errorf("failed to run postprocessing: %v", err)
	}

	fmt.Printf("Python package post-processing completed successfully\n")
	fmt.Println("Note: Now using unified version management (same as C++)")

	return nil
}

// createGitHubRelease attempts to create a GitHub Release for the package (legacy function)
func createGitHubRelease(packageName, version, currentDir string) error {
	fmt.Println("Starting GitHub Release creation...")

	// Check if we're in a GitHub Actions environment
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		fmt.Println("Not running in GitHub Actions environment, skipping GitHub Release creation")
		return nil // Non-GitHub Actions environment doesn't report error, just skip
	}

	// Check if GitHub CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not installed: %v", err)
	}

	// Check if GITHUB_REPOSITORY is set
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return fmt.Errorf("GITHUB_REPOSITORY environment variable is not set")
	}

	// Check if GITHUB_TOKEN is set
	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	fmt.Printf("Repository: %s\n", repo)
	fmt.Printf("Package: %s\n", packageName)
	fmt.Printf("Version: %s\n", version)

	// Create release tag using the extracted version
	releaseTag := fmt.Sprintf("%s/%s", packageName, version)
	fmt.Printf("Release tag: %s\n", releaseTag)

	// Use GitHub CLI to create the release
	// First check if release already exists
	fmt.Println("Checking if release already exists...")
	checkCmd := exec.Command("gh", "release", "view", releaseTag, "--repo", repo)
	checkCmd.Env = append(os.Environ(), "GITHUB_TOKEN="+os.Getenv("GITHUB_TOKEN"))

	if err := checkCmd.Run(); err == nil {
		fmt.Printf("Release %s already exists, updating existing release...\n", releaseTag)

		// Update the existing release instead of deleting
		updateCmd := exec.Command("gh", "release", "edit", releaseTag,
			"--title", fmt.Sprintf("Release %s %s", packageName, version),
			"--notes", fmt.Sprintf("Automated release for %s version %s", packageName, version),
			"--repo", repo)
		updateCmd.Env = append(os.Environ(), "GITHUB_TOKEN="+os.Getenv("GITHUB_TOKEN"))
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr

		if err := updateCmd.Run(); err != nil {
			return fmt.Errorf("failed to update existing GitHub release: %v", err)
		}

		fmt.Printf("Successfully updated GitHub Release: %s\n", releaseTag)
	} else {
		fmt.Println("Release does not exist, creating new release...")

		// Create the release using GitHub CLI
		createCmd := exec.Command("gh", "release", "create", releaseTag,
			"--title", fmt.Sprintf("Release %s %s", packageName, version),
			"--notes", fmt.Sprintf("Automated release for %s version %s", packageName, version),
			"--repo", repo,
			"--draft=false",
			"--prerelease=false")
		createCmd.Env = append(os.Environ(), "GITHUB_TOKEN="+os.Getenv("GITHUB_TOKEN"))
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr

		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create GitHub release: %v", err)
		}

		fmt.Printf("Successfully created GitHub Release: %s\n", releaseTag)
	}

	fmt.Printf("Package %s version %s is now available via: llgo get github.com/%s/%s@%s\n",
		packageName, version, repo, packageName, version)

	return nil
}

// updateLLPkgStoreJSON updates the llpkgstore.json file with new package information
// Uses C++ compatible format for version mapping
func updateLLPkgStoreJSON(packageName, pythonVersion, goVersion, jsonPath string) error {
	// Ensure the directory exists for the target file
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %v", jsonPath, err)
	}

	// Read existing JSON file if it exists
	var llpkgStore LLPkgStoreJSON
	if _, err := os.Stat(jsonPath); err == nil {
		// File exists, read it
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to read existing llpkgstore.json: %v", err)
		}

		// Check if file is empty
		if len(strings.TrimSpace(string(data))) == 0 {
			// File is empty, create new structure
			llpkgStore = LLPkgStoreJSON{
				Packages: make(map[string]PackageInfo),
			}
		} else {
			// File has content, try to parse it
			if err := json.Unmarshal(data, &llpkgStore); err != nil {
				return fmt.Errorf("failed to parse existing llpkgstore.json: %v", err)
			}
		}
	} else {
		// File doesn't exist, create new structure
		llpkgStore = LLPkgStoreJSON{
			Packages: make(map[string]PackageInfo),
		}
	}

	// Initialize package info if it doesn't exist
	if _, exists := llpkgStore.Packages[packageName]; !exists {
		llpkgStore.Packages[packageName] = PackageInfo{
			Versions: []VersionInfo{},
		}
	}

	// Check if this Python version already exists
	packageInfo := llpkgStore.Packages[packageName]
	found := false
	for i, versionInfo := range packageInfo.Versions {
		if versionInfo.Python == pythonVersion {
			// Python version exists, add Go version if not already present
			if !contains(versionInfo.Go, goVersion) {
				packageInfo.Versions[i].Go = append(versionInfo.Go, goVersion)
			}
			found = true
			break
		}
	}

	// If Python version doesn't exist, add new version entry
	if !found {
		packageInfo.Versions = append(packageInfo.Versions, VersionInfo{
			Python: pythonVersion,
			Go:     []string{goVersion},
		})
	}

	llpkgStore.Packages[packageName] = packageInfo

	// Write updated JSON back to file
	data, err := json.MarshalIndent(llpkgStore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal llpkgstore.json: %v", err)
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write llpkgstore.json: %v", err)
	}

	fmt.Printf("Updated llpkgstore.json with package %s (Python: %s -> Go: %s)\n", packageName, pythonVersion, goVersion)
	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// findLLPkgPublicPath finds the path to llpkg/public/llpkgstore.json
// It searches up the directory tree to find the llpkg repository root
func findLLPkgPublicPath(currentDir string) string {
	// Start from current directory and go up the tree
	dir := currentDir
	for {
		// Check if we're in the llpkg repository root
		publicPath := filepath.Join(dir, "public", "llpkgstore.json")
		if _, err := os.Stat(publicPath); err == nil {
			return publicPath
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return ""
}

// createGitTag creates a git tag for the specified version
func createGitTag(version, currentDir string) error {
	fmt.Printf("Creating git tag: %s\n", version)

	// Check if tag already exists
	cmd := exec.Command("git", "tag", "-l", version)
	cmd.Dir = currentDir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check existing tags: %v", err)
	}

	if strings.TrimSpace(string(output)) == version {
		fmt.Printf("Tag %s already exists, skipping creation\n", version)
		return nil
	}

	// Create the tag
	cmd = exec.Command("git", "tag", "-a", version, "-m", fmt.Sprintf("Release %s", version))
	cmd.Dir = currentDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create git tag %s: %v", version, err)
	}

	fmt.Printf("Successfully created git tag: %s\n", version)

	// Try to push the tag to remote (only if we're in a git repository with remote)
	cmd = exec.Command("git", "remote", "-v")
	cmd.Dir = currentDir
	output, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		fmt.Printf("Pushing tag %s to remote repository...\n", version)
		cmd = exec.Command("git", "push", "origin", version)
		cmd.Dir = currentDir
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to push tag to remote: %v\n", err)
			fmt.Println("Tag created locally but not pushed to remote")
		} else {
			fmt.Printf("Successfully pushed tag %s to remote repository\n", version)
		}
	} else {
		fmt.Println("No remote repository found, tag created locally only")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(postProcessingCmd)
}
