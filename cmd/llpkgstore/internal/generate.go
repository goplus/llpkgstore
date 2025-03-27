package internal

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/file"
	"github.com/goplus/llpkgstore/internal/actions/generator/llcppg"
	"github.com/goplus/llpkgstore/internal/actions/pc"
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

	temp, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(temp)
	pcName, err := uc.Installer.Install(uc.Pkg, temp)
	if err != nil {
		log.Fatal(err)
	}
	// copy file for debugging.
	file.CopyFilePattern(temp, dir, "*.pc")
	// try llcppcfg if llcppg.cfg dones't exist
	if _, err := os.Stat(filepath.Join(dir, "llcppg.cfg")); os.IsNotExist(err) {
		cmd := exec.Command("llcppcfg", pcName)
		cmd.Dir = dir
		pc.SetPath(cmd, temp)
		ret, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("llcppcfg execute fail: %s", string(ret))
		}
	}

	generator := llcppg.New(dir, cfg.Upstream.Package.Name, temp)

	if err := generator.Generate(dir); err != nil {
		log.Fatal(err)
	}
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
