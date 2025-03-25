package internal

import (
	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Verify a PR",
	Long:  ``,
	Run:   runReleaseCmd,
}

func runReleaseCmd(_ *cobra.Command, _ []string) {
	actions.NewDefaultClient().Release()
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
