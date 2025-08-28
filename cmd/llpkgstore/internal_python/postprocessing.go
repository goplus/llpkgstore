package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goplus/llpkgstore/config"
	"github.com/spf13/cobra"
)

var postProcessingCmd = &cobra.Command{
	Use:   "postprocessing",
	Short: "Process merged PR for Python packages",
	Long:  `Process merged PR for Python packages with GitHub Release support`,
	RunE:  runPythonPostProcessingCmd,
}

// LLPkgStoreJSON represents the structure of llpkgstore.json
type LLPkgStoreJSON struct {
	Packages map[string]PackageInfo `json:"packages"`
}

// PackageInfo represents package version information
type PackageInfo struct {
	Versions []VersionInfo `json:"versions"`
}

// VersionInfo represents version mapping
type VersionInfo struct {
	Python string   `json:"python"`
	Go     []string `json:"go"`
}

func runPythonPostProcessingCmd(_ *cobra.Command, _ []string) error {
	fmt.Println("Starting Python package post-processing...")

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

	cfg, err := config.ParseLLPkgConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to parse llpkg.cfg: %v", err)
	}

	// Check for generated Python package files
	requiredFiles := []string{"go.mod", "go.sum"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(currentDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file not found: %s", file)
		}
	}

	// Extract version from commit message
	version, err := extractVersionFromCommit(currentDir)
	if err != nil {
		fmt.Printf("Warning: Failed to extract version from commit: %v\n", err)
		fmt.Println("Falling back to default version v0.0.2")
		version = "v0.0.2"
	}

	pythonVersion := cfg.Upstream.Package.Version
	packageName := cfg.Upstream.Package.Name

	// Skip llpkgstore.json update for now - focus only on GitHub Release
	fmt.Println("Skipping llpkgstore.json update - focusing on GitHub Release creation")

	// Try to create GitHub Release if we're in a GitHub Actions environment
	if err := createGitHubRelease(packageName, version, currentDir); err != nil {
		fmt.Printf("Warning: Failed to create GitHub Release: %v\n", err)
		fmt.Println("This is normal if not running in GitHub Actions or if release already exists")
	}

	fmt.Printf("Python package post-processing completed successfully\n")
	fmt.Printf("Package: %s\n", packageName)
	fmt.Printf("Python Version: %s\n", pythonVersion)
	fmt.Printf("Go Version: %s\n", version)
	fmt.Println("Note: This is a simplified post-processing process for Python packages")

	return nil
}

// extractVersionFromCommit extracts version from the latest commit message
func extractVersionFromCommit(currentDir string) (string, error) {
	fmt.Println("Extracting version from commit message...")

	// 首先尝试从当前分支获取
	cmd := exec.Command("git", "log", "-5", "--pretty=format:%s")
	cmd.Dir = currentDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit messages: %v", err)
	}

	commitMessages := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Printf("Recent commit messages from current branch: %v\n", commitMessages)

	// 如果当前分支没有找到版本信息，尝试从 main 分支获取
	for _, commitMessage := range commitMessages {
		commitMessage = strings.TrimSpace(commitMessage)
		if commitMessage == "" {
			continue
		}

		// 跳过自动生成的 commit 消息
		if strings.Contains(commitMessage, "Update llpkgstore.json") {
			continue
		}

		// 尝试解析版本
		version, err := parseVersionFromCommitMessage(commitMessage)
		if err == nil {
			fmt.Printf("Found version in commit message: %s -> %s\n", commitMessage, version)
			return version, nil
		}
	}

	// 如果当前分支没有找到，尝试从 main 分支获取
	fmt.Println("No version found in current branch, trying main branch...")
	cmd = exec.Command("git", "log", "origin/main", "-5", "--pretty=format:%s")
	cmd.Dir = currentDir
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("Warning: failed to get commit messages from main branch: %v\n", err)
		return "", fmt.Errorf("no version pattern found in recent commit messages")
	}

	commitMessages = strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Printf("Recent commit messages from main branch: %v\n", commitMessages)

	for _, commitMessage := range commitMessages {
		commitMessage = strings.TrimSpace(commitMessage)
		if commitMessage == "" {
			continue
		}

		// 跳过自动生成的 commit 消息
		if strings.Contains(commitMessage, "Update llpkgstore.json") {
			continue
		}

		// 尝试解析版本
		version, err := parseVersionFromCommitMessage(commitMessage)
		if err == nil {
			fmt.Printf("Found version in main branch commit message: %s -> %s\n", commitMessage, version)
			return version, nil
		}
	}

	return "", fmt.Errorf("no version pattern found in recent commit messages")
}

// parseVersionFromCommitMessage parses version from commit message
func parseVersionFromCommitMessage(commitMessage string) (string, error) {
	// Define regex patterns for different formats
	patterns := []string{
		`Release-as:\s*[^/]+/(v[\d.]+)`, // "Release-as: numpy/v1.26.4"
		`Release-as:\s*(v[\d.]+)`,       // "Release-as: v1.26.4"
		`Release:\s*[^/]+/(v[\d.]+)`,    // "Release: numpy/v1.26.4"
		`Release:\s*(v[\d.]+)`,          // "Release: v1.26.4"
		`Version:\s*(v[\d.]+)`,          // "Version: v1.26.4"
		`version:\s*(v[\d.]+)`,          // "version: v1.26.4"
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(commitMessage)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("no version pattern found in commit message: %s", commitMessage)
}

// createGitHubRelease attempts to create a GitHub Release for the package
func createGitHubRelease(packageName, version, currentDir string) error {
	fmt.Println("Starting GitHub Release creation...")

	// Check if we're in a GitHub Actions environment
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return fmt.Errorf("not running in GitHub Actions environment")
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

	fmt.Printf("Repository: %s\n", repo)
	fmt.Printf("Package: %s\n", packageName)
	fmt.Printf("Version: %s\n", version)

	// Create release tag using the extracted version
	releaseTag := fmt.Sprintf("%s/%s", packageName, version)
	fmt.Printf("Release tag: %s\n", releaseTag)

	// Use GitHub CLI to create the release
	// First check if release already exists
	fmt.Println("Checking if release already exists...")
	checkCmd := fmt.Sprintf("gh release view %s --repo %s >/dev/null 2>&1", releaseTag, repo)
	if err := exec.Command("bash", "-c", checkCmd).Run(); err == nil {
		fmt.Printf("Release %s already exists, deleting existing release...\n", releaseTag)

		// Delete the existing release
		deleteCmd := fmt.Sprintf("gh release delete %s --repo %s --yes", releaseTag, repo)
		fmt.Printf("Executing delete command: %s\n", deleteCmd)

		deleteCmdExec := exec.Command("bash", "-c", deleteCmd)
		deleteCmdExec.Stdout = os.Stdout
		deleteCmdExec.Stderr = os.Stderr
		deleteCmdExec.Env = append(os.Environ(), "GITHUB_TOKEN="+os.Getenv("GITHUB_TOKEN"))

		if err := deleteCmdExec.Run(); err != nil {
			return fmt.Errorf("failed to delete existing GitHub release: %v", err)
		}

		fmt.Printf("Successfully deleted existing release: %s\n", releaseTag)
	} else {
		fmt.Println("Release does not exist, will create new release...")
	}

	fmt.Println("Creating new release...")

	// Create the release using GitHub CLI
	createCmd := fmt.Sprintf("gh release create %s --title 'Release %s %s' --notes 'Automated release for %s version %s' --repo %s --draft=false --prerelease=false",
		releaseTag, packageName, version, packageName, version, repo)

	fmt.Printf("Executing command: %s\n", createCmd)

	cmd := exec.Command("bash", "-c", createCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GITHUB_TOKEN="+os.Getenv("GITHUB_TOKEN"))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub release: %v", err)
	}

	fmt.Printf("Successfully created GitHub Release: %s\n", releaseTag)
	fmt.Printf("Package %s version %s is now available via: llgo get github.com/%s/%s@%s\n",
		packageName, version, repo, packageName, version)

	return nil
}

// updateLLPkgStoreJSON updates the llpkgstore.json file with new package information
func updateLLPkgStoreJSON(packageName, pythonVersion, goVersion string) error {
	jsonPath := "llpkgstore.json"

	// Read existing JSON file if it exists
	var llpkgStore LLPkgStoreJSON
	if _, err := os.Stat(jsonPath); err == nil {
		// File exists, read it
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to read existing llpkgstore.json: %v", err)
		}

		if err := json.Unmarshal(data, &llpkgStore); err != nil {
			return fmt.Errorf("failed to parse existing llpkgstore.json: %v", err)
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

func init() {
	rootCmd.AddCommand(postProcessingCmd)
}
