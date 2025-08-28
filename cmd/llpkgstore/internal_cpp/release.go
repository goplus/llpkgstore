package internal

import (
	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Build and upload binary packages",
	Long:  ``,
	RunE:  runReleaseCmd,
}

func runReleaseCmd(_ *cobra.Command, _ []string) error {
	client, err := actions.NewDefaultClient()
	if err != nil {
		return err
	}
	return client.Release()
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
