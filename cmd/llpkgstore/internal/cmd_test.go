package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goplus/llpkgstore/internal/file"
)

func TestCMD(t *testing.T) {
	// ../../../_demo
	demoDir := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(currentDir()))), "_demo")
	runLLCppgGenerateWithDir(demoDir)

	// remove go.mod
	file.RemovePattern(filepath.Join(demoDir, "go.*"))
	os.Chdir(filepath.Dir(demoDir))

	file.CopyFilePattern(demoDir, filepath.Dir(demoDir), "*.pc")
	runDemotestCmd(nil, nil)
}
