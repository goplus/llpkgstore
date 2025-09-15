package main

import (
	"fmt"
	"os"
	"path/filepath"

	cmd_cpp "github.com/goplus/llpkgstore/cmd/llpkgstore/internal/cpp"
	cmd_python "github.com/goplus/llpkgstore/cmd/llpkgstore/internal/python"
	"github.com/goplus/llpkgstore/config"
)

// detectPackageType detects the package type in the current directory or specified directory
func detectPackageType(dir string) (string, error) {
	// Check if this is a postprocessing command - skip config file check
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "postprocessing" {
			// For postprocessing, return a default type and let the command handle it
			return "python", nil
		}
	}

	// If no directory is specified, use the current directory
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %v", err)
		}
	}

	// Find llpkg.cfg file
	cfgPath := filepath.Join(dir, "llpkg.cfg")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return "", fmt.Errorf("llpkg.cfg file not found in directory %s", dir)
	}

	// Use config.ParseLLPkgConfig to read and parse configuration file
	cfg, err := config.ParseLLPkgConfig(cfgPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse configuration file: %v", err)
	}

	// If type field is empty or not present, default to cpp
	if cfg.Type == "" {
		return "cpp", nil
	}

	return cfg.Type, nil
}

// findLLPkgConfigDir finds the directory containing llpkg.cfg
func findLLPkgConfigDir() (string, error) {
	// Check if this is a postprocessing command - skip config file check
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "postprocessing" {
			// For postprocessing, let the command handle config file discovery
			return "", nil
		}
	}

	// Check if there are directory paths in command line arguments
	for _, arg := range args {
		// Skip flag arguments
		if arg[0] == '-' {
			continue
		}

		// Check if it's a directory path
		if stat, err := os.Stat(arg); err == nil && stat.IsDir() {
			if _, err := os.Stat(filepath.Join(arg, "llpkg.cfg")); err == nil {
				return arg, nil
			}
		}
	}

	// If not found, check current directory
	if _, err := os.Stat("llpkg.cfg"); err == nil {
		return "", nil // Empty string represents current directory
	}

	return "", fmt.Errorf("directory containing llpkg.cfg not found")
}

func main() {
	// Check if help flag is provided
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		// Show help for Python version by default
		fmt.Println("llpkgstore - Package management tool for Python and C/C++ packages")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  llpkgstore [command] [flags]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  generate      Generate Go bindings for Python packages")
		fmt.Println("  verification  Verify generated packages")
		fmt.Println("  postprocessing Process merged PR for Python packages")
		fmt.Println("  release       Build and upload Python binary packages")
		fmt.Println("  install       Install Python packages")
		fmt.Println("  test          Test Python verification functionality")
		fmt.Println("")
		fmt.Println("Flags:")
		fmt.Println("  -h, --help    Show this help message")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  llpkgstore generate")
		fmt.Println("  llpkgstore verification")
		fmt.Println("  llpkgstore postprocessing")
		return
	}

	// Find directory containing llpkg.cfg
	configDir, err := findLLPkgConfigDir()
	if err != nil {
		// If configuration file not found, show help and exit
		fmt.Printf("Error: %v\n", err)
		fmt.Println("")
		fmt.Println("Please ensure you are in a directory containing llpkg.cfg file.")
		fmt.Println("Use 'llpkgstore --help' for more information.")
		return
	}

	// Detect package type
	packageType, err := detectPackageType(configDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Detected package type: %s\n", packageType)

	// Select appropriate command implementation based on package type
	switch packageType {
	case "python":
		fmt.Println("Using Python version of llpkgstore command")
		cmd_python.Execute()
	case "cpp":
		fmt.Println("Using C++ version of llpkgstore command")
		cmd_cpp.Execute()
	default:
		fmt.Printf("Error: Currently only python and c/c++ packages are supported, detected type: %s\n", packageType)
		return
	}
}
