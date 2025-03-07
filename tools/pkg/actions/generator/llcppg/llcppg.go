package llcppg

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goplus/llpkg/tools/pkg/actions/generator"
)

var (
	goModFile         = "go.mod"
	ErrLLCppgGenerate = errors.New("llcppg: cannot generate: ")
)

const (
	// llcppg default config file, which MUST exist in specifed dir
	llcppgConfigFile = "llcppg.cfg"
)

// llcppgGenerator implements Generator interface, which use llcppg tool to generate llpkg.
type llcppgGenerator struct {
	dir string // llcppg.cfg abs path
}

func New(dir string) generator.Generator {
	return &llcppgGenerator{dir: dir}
}

func (l *llcppgGenerator) findGoMod() (baseDir string) {
	filepath.Walk(l.dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == goModFile {
			baseDir = filepath.Dir(path)
			return filepath.SkipAll
		}
		return nil
	})
	return
}

func (l *llcppgGenerator) Generate() error {
	cmd := exec.Command("llcppg", llcppgConfigFile)
	cmd.Dir = l.dir
	ret, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Join(ErrLLCppgGenerate, errors.New(string(ret)))
	}
	baseDir := l.findGoMod()
	if baseDir == "" {
		return errors.Join(ErrLLCppgGenerate, errors.New("dir not found"))
	}
	// copy out
	if err := os.CopyFS(l.dir, os.DirFS(baseDir)); err != nil {
		return errors.Join(ErrLLCppgGenerate, err)
	}
	// clean path
	os.RemoveAll(baseDir)
	// remove llcppg.symb.json
	os.Remove(filepath.Join(l.dir, "llcppg.symb.json"))
	return nil
}
