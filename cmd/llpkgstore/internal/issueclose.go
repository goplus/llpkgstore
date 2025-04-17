package internal

import (
	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var issueCloseCmd = &cobra.Command{
	Use:   "issueclose",
	Short: "Legacy version maintenance on label creating",
	Long:  ``,

	RunE: runIssueCloseCmd,
}

func runIssueCloseCmd(cmd *cobra.Command, args []string) error {
	client, err := actions.NewDefaultClient()
	if err != nil {
		return err
	}
	return client.CleanResource()
}

func init() {
	rootCmd.AddCommand(issueCloseCmd)
}
