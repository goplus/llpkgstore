package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/goplus/llpkgstore/config"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [LLPkgConfigFilePath]",
	Short: "Manually install a Python package",
	Long:  `Manually install a Python package from llpkg.cfg file using pip.`,
	Args:  cobra.ExactArgs(1),
	RunE:  manuallyInstall,
}

func manuallyInstall(cmd *cobra.Command, args []string) error {
	cfgPath := args[0]

	// Check if configuration file exists
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", cfgPath)
	}

	// Parse configuration file
	LLPkgConfig, err := config.ParseLLPkgConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to parse configuration file: %v", err)
	}

	// Check package type
	if LLPkgConfig.Type != "python" {
		return fmt.Errorf("unsupported package type: %s, currently only Python packages are supported", LLPkgConfig.Type)
	}

	// Get output directory
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	// If output directory is empty, use current directory
	if output == "" {
		output = "."
	}

	// Ensure output directory exists
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	log.Printf("Starting to install Python package: %s==%s", LLPkgConfig.Upstream.Package.Name, LLPkgConfig.Upstream.Package.Version)
	log.Printf("Output directory: %s", output)

	// Create upstream instance
	upstream, err := config.NewUpstreamFromConfig(LLPkgConfig.Upstream)
	if err != nil {
		return fmt.Errorf("failed to create upstream instance: %v", err)
	}

	// Execute installation
	installedPackages, err := upstream.Installer.Install(upstream.Pkg, output)
	if err != nil {
		return fmt.Errorf("installation failed: %v", err)
	}

	log.Printf("Installation successful! Installed packages: %v", installedPackages)

	// Display installation results
	fmt.Printf("✓ Successfully installed Python package: %s==%s\n", LLPkgConfig.Upstream.Package.Name, LLPkgConfig.Upstream.Package.Version)
	fmt.Printf("  Installation location: %s\n", output)
	fmt.Printf("  Installed packages: %v\n", installedPackages)

	// If it's a pip installer, show additional information
	if LLPkgConfig.Upstream.Installer.Name == "pip" {
		fmt.Println("\nNote:")
		fmt.Println("- Package has been installed to the specified directory via pip3")
		fmt.Println("- You can use 'generate' command to generate Go bindings")
		fmt.Println("- You can use 'test' command to verify installation results")
	}

	return nil
}

func init() {
	installCmd.Flags().StringP("output", "o", "", "Installation output directory (default: current directory)")
	rootCmd.AddCommand(installCmd)
}
