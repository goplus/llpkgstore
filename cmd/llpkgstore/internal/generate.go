package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions/generator/llcppg"
	"github.com/goplus/llpkgstore/internal/file"
	"github.com/goplus/llpkgstore/internal/pc"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "PR Verification",
	Long:  ``,
	RunE:  runLLCppgGenerate,
}

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func runLLCppgGenerateWithDir(dir string) error {
	cfg, err := config.ParseLLPkgConfig(filepath.Join(dir, LLGOModuleIdentifyFile))
	if err != nil {
		return fmt.Errorf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		return err
	}
	log.Printf("Start to generate %s", uc.Pkg.Name)

	tempDir, err := os.MkdirTemp("", "llpkg-tool")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	pcName, err := uc.Installer.Install(uc.Pkg, tempDir)
	if err != nil {
		return err
	}
	// copy file for debugging.
	err = file.CopyFilePattern(tempDir, dir, "*.pc")
	if err != nil {
		return err
	}
	// try llcppcfg if llcppg.cfg dones't exist
	if _, err := os.Stat(filepath.Join(dir, "llcppg.cfg")); os.IsNotExist(err) {
		cmd := exec.Command("llcppcfg", pcName[0])
		cmd.Dir = dir
		pc.SetPath(cmd, tempDir)
		ret, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("llcppcfg execute fail: %s", string(ret))
		}
	}

	generator := llcppg.New(dir, cfg.Upstream.Package.Name, tempDir)

	return generator.Generate(dir)
}

func runLLCppgGenerate(_ *cobra.Command, args []string) error {
	exec.Command("conan", "profile", "detect").Run()

	path := currentDir()
	// by default, use current dir
	if len(args) == 0 {
		return runLLCppgGenerateWithDir(path)
	}
	for _, argPath := range args {
		absPath, err := filepath.Abs(argPath)
		if err != nil {
			continue
		}
		err = runLLCppgGenerateWithDir(absPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
