package internal

import (
	"strings"

	"github.com/goplus/llpkgstore/config"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [LLPkgConfigFilePath]",
	Short: "Manually install a package",
	Long:  `Manually install a package from cfg file.`,
	Args:  cobra.ExactArgs(1),
	Run:   manuallyInstall,
}

func manuallyInstall(cmd *cobra.Command, args []string) {
	cfgPath := strings.Join(args, " ")
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		cmd.PrintErrln("Error retrieving 'output' flag:", err)
		return
	}
	LLPkgConfig, err := config.ParseLLPkgConfig(cfgPath)
	if err != nil {
		cmd.PrintErrln("Error parsing LLPkgConfig:", err)
		return
	}
	upstream, err := config.NewUpstreamFromConfig(LLPkgConfig.Upstream)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	upstream.Installer.Install(upstream.Pkg, output)
}

func init() {
	installCmd.Flags().StringP("output", "o", "", "Path to the output file")
	installCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(installCmd)
}
