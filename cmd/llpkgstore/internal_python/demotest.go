package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// demotestCmd represents the demotest command for Python packages
var demotestCmd = &cobra.Command{
	Use:   "demotest",
	Short: "A tool that runs all Python package demos",
	Long:  `A tool that runs all demo tests for Python packages to verify the generated Go bindings work correctly.`,
	RunE:  runPythonDemotestCmd,
}

func runPythonDemotestCmd(cmd *cobra.Command, args []string) error {
	var paths []string
	pathEnv := os.Getenv("LLPKG_PATH")
	if pathEnv != "" {
		json.Unmarshal([]byte(pathEnv), &paths)
	} else {
		// not in github action
		paths = append(paths, currentDir())
	}

	for _, path := range paths {
		if err := runPythonDemo(path); err != nil {
			return err
		}
	}
	return nil
}

func runPythonDemo(demoRoot string) error {
	demosPath := filepath.Join(demoRoot, "_demo")

	fmt.Printf("Testing Python demos in %s\n", demosPath)
	
	// Check if _demo directory exists
	if _, err := os.Stat(demosPath); os.IsNotExist(err) {
		return fmt.Errorf("demotest: demo directory not found: %s", demosPath)
	}

	// Read and run all demos
	demos, err := os.ReadDir(demosPath)
	if err != nil {
		return fmt.Errorf("demotest: failed to read demo directory: %w", err)
	}

	for _, demo := range demos {
		if demo.IsDir() {
			fmt.Printf("Running Python demo: %s\n", demo.Name())
			if demoErr := runPythonCommand(demoRoot, filepath.Join(demosPath, demo.Name()), "llgo", "run", "."); demoErr != nil {
				return fmt.Errorf("demotest: failed to run Python demo: %s: %w", demo.Name(), demoErr)
			}
		}
	}
	return nil
}

func runPythonCommand(pcPath, dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Set environment variables for Python packages
	cmd.Env = append(os.Environ(), "PYTHONPATH="+pcPath)
	
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(demotestCmd)
} 