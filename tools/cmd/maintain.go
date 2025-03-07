package cmd

import "github.com/spf13/cobra"

var maintainCmd = &cobra.Command{
	Use:   "post-processing",
	Short: "Verify a PR",
	Long:  ``,
	Run:   runMaintainCmd,
}

func runMaintainCmd(_ *cobra.Command, _ []string) {
	//TODO
}

func init() {
	rootCmd.AddCommand(maintainCmd)
}
