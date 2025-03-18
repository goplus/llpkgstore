package internal

import (
	"github.com/goplus/llpkgstore/actions"
	"github.com/spf13/cobra"
)

var (
	labelName string
)

var labelCreateCmd = &cobra.Command{
	Use:   "labelcreate",
	Short: "Legacy version maintenance on label creating",
	Long:  ``,

	Run: runLabelCreateCmd,
}

func runLabelCreateCmd(cmd *cobra.Command, args []string) {
	if labelName == "" {
		panic("no label name")
	}
	actions.NewDefaultClient().CreateBranchFromLabel(labelName)
}

func init() {
	labelCreateCmd.Flags().StringVarP(&labelName, "label", "l", "", "input the created label name")
	rootCmd.AddCommand(labelCreateCmd)
}
