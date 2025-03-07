package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/goplus/llpkg/tools/pkg/actions/generator/llcppg"
	"github.com/goplus/llpkg/tools/pkg/config"
	"github.com/spf13/cobra"
)

const LLGOModuleIdentifyFile = "llpkg.cfg"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a llpkg",
	Long:  ``,
	Run:   runLLCppgGenerate,
}

func currentDir() string {
	absFile, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot get current path: %v", err)
	}
	return filepath.Dir(absFile)
}

func runLLCppgGenerateWithDir(dir string) {
	cfg, err := config.ParseLLpkgConfig(filepath.Join(dir, LLGOModuleIdentifyFile))
	if err != nil {
		log.Fatalf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.UpstreamConfig)
	if err != nil {
		log.Fatal()
	}
	err = uc.Installer().Install(uc.Package(), dir)
	if err != nil {
		log.Fatal(err)
	}
	// we have to feed the pc to llcppg
	os.Setenv("PKG_CONFIG_PATH", dir)

	err = llcppg.New(dir).Generate()
	if err != nil {
		log.Fatal(err)
	}
}

func runLLCppgGenerate(_ *cobra.Command, _ []string) {
	runLLCppgGenerateWithDir(currentDir())
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
