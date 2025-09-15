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
	RunE:  manuallyInstall,
}

func manuallyInstall(cmd *cobra.Command, args []string) error {
	cfgPath := strings.Join(args, " ")
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	LLPkgConfig, err := config.ParseLLPkgConfig(cfgPath)
	if err != nil {
		return err
	}
	upstream, err := config.NewUpstreamFromConfig(LLPkgConfig.Upstream)
	if err != nil {
		return err
	}
	_, err = upstream.Installer.Install(upstream.Pkg, output)
	return err
}

func init() {
	installCmd.Flags().StringP("output", "o", "", "Path to the output file")
	installCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(installCmd)
}
