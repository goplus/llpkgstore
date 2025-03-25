package internal

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

<<<<<<< HEAD
	"github.com/MeteorsLiu/llpkgstore/config"
	"github.com/MeteorsLiu/llpkgstore/internal/actions/generator/llcppg"
=======
	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/generator/llcppg"
>>>>>>> 6ac1bf45c9b40c79c28d41e005b9bbd7259b7688
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "PR Verification",
	Long:  ``,
	Run:   runLLCppgGenerate,
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func removePattern(pattern string) {
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		os.Remove(match)
	}
}

func runLLCppgGenerateWithDir(dir string) {
	cfg, err := config.ParseLLPkgConfig(filepath.Join(dir, LLGOModuleIdentifyFile))
	if err != nil {
		log.Fatalf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Start to generate %s", uc.Pkg.Name)
	_, err = uc.Installer.Install(uc.Pkg, dir)
	if err != nil {
		log.Fatal(err)
	}
	// we have to feed the pc to llcppg
	os.Setenv("PKG_CONFIG_PATH", dir)

	// try llcppcfg if llcppg.cfg dones't exist
	if _, err := os.Stat(filepath.Join(dir, "llcppg.cfg")); os.IsNotExist(err) {
		cmd := exec.Command("llcppcfg", uc.Pkg.Name)
		cmd.Dir = dir

		ret, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("llcppcfg execute fail: %s", string(ret))
		}
	}

	generator := llcppg.New(dir, cfg.Upstream.Package.Name)

	if err := generator.Generate(dir); err != nil {
		log.Fatal(err)
	}

	removePattern("*.sh")
	removePattern("*.bat")
}

func runLLCppgGenerate(_ *cobra.Command, args []string) {
	exec.Command("conan", "profile", "detect").Run()

	path := currentDir()
	// by default, use current dir
	if len(args) == 0 {
		runLLCppgGenerateWithDir(path)
		return
	}
	for _, argPath := range args {
		absPath, err := filepath.Abs(argPath)
		if err != nil {
			continue
		}
		runLLCppgGenerateWithDir(absPath)
	}

}

func init() {
	rootCmd.AddCommand(generateCmd)
}
