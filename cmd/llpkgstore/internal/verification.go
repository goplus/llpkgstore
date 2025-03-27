package internal

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goplus/llpkgstore/config"
	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/goplus/llpkgstore/internal/actions/generator/llcppg"
	"github.com/spf13/cobra"
)

const LLGOModuleIdentifyFile = "llpkg.cfg"

var verificationCmd = &cobra.Command{
	Use:   "verification",
	Short: "PR Verification",
	Long:  ``,
	Run:   runLLCppgVerification,
}

func runLLCppgVerificationWithDir(dir string) {
	cfg, err := config.ParseLLPkgConfig(filepath.Join(dir, LLGOModuleIdentifyFile))
	if err != nil {
		log.Fatalf("parse config error: %v", err)
	}
	uc, err := config.NewUpstreamFromConfig(cfg.Upstream)
	if err != nil {
		log.Fatal(err)
	}
	_, err = uc.Installer.Install(uc.Pkg, dir)
	if err != nil {
		log.Fatal(err)
	}
	generator := llcppg.New(dir, cfg.Upstream.Package.Name, dir)

	generated := filepath.Join(dir, ".generated")
	os.Mkdir(generated, 0777)
	defer os.Remove(generated)

	if err := generator.Generate(generated); err != nil {
		log.Fatal(err)
	}
	if err := generator.Check(generated); err != nil {
		log.Fatal(err)
	}
}

func runLLCppgVerification(_ *cobra.Command, _ []string) {
	exec.Command("conan", "profile", "detect").Run()

	paths := actions.NewDefaultClient().CheckPR()

	for _, path := range paths {
		absPath, _ := filepath.Abs(path)
		runLLCppgVerificationWithDir(absPath)
	}
	// output parsed path to Github Env for demotest
	b, _ := json.Marshal(&paths)
	actions.Setenv(map[string]string{
		"LLPKG_PATH": string(b),
	})
}

func init() {
	rootCmd.AddCommand(verificationCmd)
}
