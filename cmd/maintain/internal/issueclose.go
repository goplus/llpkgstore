package internal

import (
	"github.com/goplus/llpkgstore/actions"
	"github.com/spf13/cobra"
)

var issueCloseCmd = &cobra.Command{
	Use:   "issueclose",
	Short: "Legacy version maintenance on label creating",
	Long:  ``,

	Run: runIssueCloseCmd,
}

func runIssueCloseCmd(cmd *cobra.Command, args []string) {
	if labelName == "" {
		panic("no label name")
	}
	actions.NewDefaultClient().CreateBranchFromLabel(labelName)
}

func init() {
	rootCmd.AddCommand(issueCloseCmd)
}
