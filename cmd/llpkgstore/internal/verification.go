package internal

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goplus/llpkgstore/internal/actions"
	"github.com/goplus/llpkgstore/internal/actions/env"
	"github.com/goplus/llpkgstore/internal/actions/generator/llcppg"
	"github.com/goplus/llpkgstore/internal/actions/llpkg"
	"github.com/spf13/cobra"
)

const LLGOModuleIdentifyFile = "llpkg.cfg"

var verificationCmd = &cobra.Command{
	Use:   "verification",
	Short: "PR Verification",
	Long:  ``,
	RunE:  runLLCppgVerification,
}

func runLLCppgVerificationWithDir(dir string) error {
	pkg, err := llpkg.NewLLPkg(dir)
	if err != nil {
		return err
	}
	uc, err := pkg.Upstream()
	if err != nil {
		return err
	}
	generator := llcppg.New(dir, pkg.Name(), dir)

	generated := filepath.Join(dir, ".generated")
	os.Mkdir(generated, 0777)

	if err := generator.Generate(generated); err != nil {
		return err
	}
	if err := generator.Check(generated); err != nil {
		return err
	}
	// TODO(ghl): upload generated result to artifact for debugging.
	os.RemoveAll(generated)
	// start prebuilt check
	_, _, err = actions.BuildBinaryZip(uc)
	return err
}

func runLLCppgVerification(_ *cobra.Command, _ []string) error {
	exec.Command("conan", "profile", "detect").Run()

	client, err := actions.NewDefaultClient()
	if err != nil {
		return err
	}
	paths, err := client.CheckPR()
	if err != nil {
		return err
	}

	for _, path := range paths {
		absPath, _ := filepath.Abs(path)
		err := runLLCppgVerificationWithDir(absPath)
		if err != nil {
			return err
		}
	}
	// output parsed path to Github Env for demotest
	b, err := json.Marshal(&paths)
	if err != nil {
		return err
	}
	return env.Setenv(env.Env{
		"LLPKG_PATH": string(b),
	})
}

func init() {
	rootCmd.AddCommand(verificationCmd)
}
