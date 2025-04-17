package internal

import (
	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var postProcessingCmd = &cobra.Command{
	Use:   "postprocessing",
	Short: "Verify a PR",
	Long:  ``,
	RunE:  runPostProcessingCmd,
}

func runPostProcessingCmd(_ *cobra.Command, _ []string) error {
	client, err := actions.NewDefaultClient()
	if err != nil {
		return err
	}
	return client.Postprocessing()
}

func init() {
	rootCmd.AddCommand(postProcessingCmd)
}
