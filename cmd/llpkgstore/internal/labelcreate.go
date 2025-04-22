package internal

import (
	"fmt"

	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/spf13/cobra"
)

var (
	labelName string
)

var labelCreateCmd = &cobra.Command{
	Use:   "labelcreate",
	Short: "Legacy version maintenance on label creating",
	Long:  ``,

	RunE: runLabelCreateCmd,
}

func runLabelCreateCmd(cmd *cobra.Command, args []string) error {
	if labelName == "" {
		return fmt.Errorf("no label name")
	}
	client, err := actions.NewDefaultClient()
	if err != nil {
		return err
	}
	return client.CreateBranchFromLabel(labelName)
}

func init() {
	labelCreateCmd.Flags().StringVarP(&labelName, "label", "l", "", "input the created label name")
	rootCmd.AddCommand(labelCreateCmd)
}
