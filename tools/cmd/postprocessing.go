package cmd

import (
	"github.com/spf13/cobra"
)

var postProcessingCmd = &cobra.Command{
	Use:   "post-processing",
	Short: "Verify a PR",
	Long:  ``,
	Run:   runPostProcessingCmd,
}

func runPostProcessingCmd(_ *cobra.Command, _ []string) {
	//TODO
}

func init() {
	rootCmd.AddCommand(postProcessingCmd)
}
