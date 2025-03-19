package internal

import (
	"github.com/MeteorsLiu/llpkgstore/actions"
	"github.com/spf13/cobra"
)

var postProcessingCmd = &cobra.Command{
	Use:   "postprocessing",
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
