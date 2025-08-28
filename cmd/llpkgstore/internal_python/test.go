package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test Python verification functionality",
	Long:  `Test the verification and check functionality for Python packages`,
	RunE:  runTest,
}

func runTest(_ *cobra.Command, args []string) error {
	// Determine test directory
	testDir := "."
	if len(args) > 0 {
		if absPath, err := filepath.Abs(args[0]); err == nil {
			testDir = absPath
		}
	}

	// Check if test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		return fmt.Errorf("test directory does not exist: %s", testDir)
	}

	// Check if llpkg.cfg file exists
	cfgPath := filepath.Join(testDir, LLGOModuleIdentifyFile)
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", cfgPath)
	}

	fmt.Printf("Starting test directory: %s\n", testDir)

	// Run verification test
	err := TestVerification(testDir)
	if err != nil {
		return fmt.Errorf("verification test failed: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
