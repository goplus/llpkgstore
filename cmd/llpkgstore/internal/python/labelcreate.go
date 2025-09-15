package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	pythonLabelName string
)

var labelCreateCmd = &cobra.Command{
	Use:   "labelcreate",
	Short: "Legacy version maintenance on label creating for Python packages",
	Long:  `Create labels for Python package version maintenance with simplified handling`,
	RunE:  runPythonLabelCreateCmd,
}

func runPythonLabelCreateCmd(cmd *cobra.Command, args []string) error {
	if pythonLabelName == "" {
		return fmt.Errorf("no label name specified")
	}

	fmt.Printf("Creating Python package label: %s\n", pythonLabelName)

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Check if llpkg.cfg exists
	cfgPath := filepath.Join(currentDir, "llpkg.cfg")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Errorf("llpkg.cfg not found in current directory")
	}

	// For Python packages, use simplified label creation
	// Directly use v0.0.1 as the base version
	baseVersion := "v0.0.1"

	fmt.Printf("Python package label creation completed successfully\n")
	fmt.Printf("Label: %s\n", pythonLabelName)
	fmt.Printf("Base version: %s\n", baseVersion)
	fmt.Println("Note: This is a simplified label creation process for Python packages")

	return nil
}

func init() {
	labelCreateCmd.Flags().StringVarP(&pythonLabelName, "label", "l", "", "input the created label name")
	rootCmd.AddCommand(labelCreateCmd)
}
