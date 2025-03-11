package llcppg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/actions/generator"
	"github.com/goplus/llpkgstore/actions/hashutils"
)

var (
	canHashFile = map[string]struct{}{
		"llcppg.pub": {},
		"go.mod":     {},
		"go.sum":     {},
	}
	ErrLlcppgGenerate = errors.New("llcppg: cannot generate: ")
	ErrLlcppgCheck    = errors.New("llcppg: check fail: ")
)

const (
	// llcppg default config file, which MUST exist in specifed dir
	llcppgConfigFile = "llcppg.cfg"
)

// canHash check file is hashable.
// Hashable file: *.go / llcppg.pub / *.symb.json
func canHash(fileName string) bool {
	if strings.Contains(fileName, ".go") {
		return true
	}
	_, ok := canHashFile[fileName]
	return ok
}

// llcppgGenerator implements Generator interface, which use llcppg tool to generate llpkg.
type llcppgGenerator struct {
	dir         string // llcppg.cfg abs path
	packageName string
}

func New(dir, packageName string) generator.Generator {
	return &llcppgGenerator{dir: dir, packageName: packageName}
}

func (l *llcppgGenerator) Generate() error {
	cmd := exec.Command("llcppg", llcppgConfigFile)
	cmd.Dir = l.dir
	ret, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Join(ErrLlcppgGenerate, errors.New(string(ret)))
	}
	return nil
}

func (l *llcppgGenerator) Check() error {
	// 1. llcppg will output to {CurrentDir}/{PackageName}
	baseDir := filepath.Join(l.dir, l.packageName)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return errors.Join(ErrLlcppgCheck, errors.New("generate fail"))
	}
	// 2. compute hash
	generated, err := hashutils.Dir(baseDir, canHash)
	if err != nil {
		return errors.Join(ErrLlcppgCheck, err)
	}
	userGenerated, err := hashutils.Dir(l.dir, canHash)
	if err != nil {
		return errors.Join(ErrLlcppgCheck, err)
	}
	// 3. check hash
	for name, hash := range userGenerated {
		generatedHash, ok := generated[name]
		if !ok {
			// if this file is hashable, it's unexpected
			// if not, we can skip it safely.
			if canHash(name) {
				return errors.Join(ErrLlcppgCheck, fmt.Errorf("unexpected file: %s", name))
			}
			// skip file
			continue
		}
		if !bytes.Equal(hash, generatedHash) {
			return errors.Join(ErrLlcppgCheck, fmt.Errorf("file not equal: %s", name))
		}
	}
	// 4. check missing file
	for name := range generated {
		if _, ok := userGenerated[name]; !ok {
			return errors.Join(ErrLlcppgCheck, fmt.Errorf("missing file: %s", name))
		}
	}
	return nil
}
