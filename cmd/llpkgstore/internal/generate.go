package internal

import (
	"log"
	"os"
	"path/filepath"

	"github.com/goplus/llpkgstore/actions"
	"github.com/goplus/llpkgstore/actions/generator/llcppg"
	"github.com/goplus/llpkgstore/config"
	"github.com/spf13/cobra"
)

const LLGOModuleIdentifyFile = "llpkg.cfg"

var generateCmd = &cobra.Command{
	Use:   "verfication",
	Short: "PR Verification",
	Long:  ``,
	Run:   runLLCppgGenerate,
}

func runLLCppgGenerateWithDir(dir string) {
	cfg, err := config.ParseLLPkgConfig(filepath.Join(dir, LLGOModuleIdentifyFile))
	if err != nil {
		log.Fatalf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		log.Fatal()
	}
	err = uc.Installer.Install(uc.Pkg, dir)
	if err != nil {
		log.Fatal(err)
	}
	// we have to feed the pc to llcppg
	os.Setenv("PKG_CONFIG_PATH", dir)

	generator := llcppg.New(dir, cfg.Upstream.Package.Name)

	if err := generator.Generate(); err != nil {
		log.Fatal(err)
	}
	if err := generator.Check(); err != nil {
		log.Fatal(err)
	}
}

func runLLCppgGenerate(_ *cobra.Command, _ []string) {
	paths := actions.NewDefaultClient().CheckPR()

	for _, path := range paths {
		absPath, _ := filepath.Abs(path)
		runLLCppgGenerateWithDir(absPath)
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
