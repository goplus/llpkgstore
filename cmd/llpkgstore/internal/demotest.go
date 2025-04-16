package internal

import (
	"encoding/json"
	"os"

	"github.com/goplus/llpkgstore/internal/demo"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var runCmd = &cobra.Command{
	Use:   "demotest",
	Short: "A tool that runs all demo",
	Run:   runDemotestCmd,
}

func runDemotestCmd(cmd *cobra.Command, args []string) {
	var paths []string
	pathEnv := os.Getenv("LLPKG_PATH")
	if pathEnv != "" {
		json.Unmarshal([]byte(pathEnv), &paths)
	} else {
		// not in github action
		paths = append(paths, currentDir())
	}

	for _, path := range paths {
		demo.Run(path)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
