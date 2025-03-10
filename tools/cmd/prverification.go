package cmd

import (
	"github.com/goplus/llpkgstore/tools/pkg/actions"
	"github.com/spf13/cobra"
)

var prverificationCmd = &cobra.Command{
	Use:   "pr-verfication",
	Short: "Verify a PR",
	Long:  ``,
	Run:   runPRVerification,
}

func runPRVerification(_ *cobra.Command, _ []string) {
	actions.NewDefaultClient().CheckPR()
}

func init() {
	rootCmd.AddCommand(prverificationCmd)
}
