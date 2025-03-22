package internal

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// rootCmd represents the base command when called without any subcommands
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A tool that runs all demo",
	Run: func(cmd *cobra.Command, args []string) {
		var paths []string
		pathEnv := os.Getenv("LLPKG_PATH")
		if pathEnv != "" {
			json.Unmarshal([]byte(pathEnv), &paths)
		} else {
			// not in github action
			paths = append(paths, currentDir())
		}

		for _, path := range paths {
			RunAllGenPkgDemos(path)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
