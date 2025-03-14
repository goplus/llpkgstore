package internal

import (
	"github.com/goplus/llpkgstore/actions"
	"github.com/spf13/cobra"
)

var postProcessingCmd = &cobra.Command{
	Use:   "post-processing",
	Short: "Verify a PR",
	Long:  ``,
	Run:   runPostProcessingCmd,
}

func runPostProcessingCmd(_ *cobra.Command, _ []string) {
	actions.NewDefaultClient().Release()
}

func init() {
	rootCmd.AddCommand(postProcessingCmd)
}
