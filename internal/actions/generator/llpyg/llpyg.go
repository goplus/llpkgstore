package llpyg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/internal/actions/generator"
	"github.com/goplus/llpkgstore/internal/file"
	"github.com/goplus/llpkgstore/internal/hashutils"
)

var (
	ErrLLPygGenerate = errors.New("llpyg: cannot generate: ")
	ErrLLPygCheck    = errors.New("llpyg: check fail: ")
)

const (
	// default llpkg repo
	goplusRepo = "github.com/goplus/llpkg/"
	// llpyg running default version
	llpygGoVersion = "1.20.14"
	// llpyg default config file, which MUST exist in specified dir
	llpygConfigFile = "llpyg.cfg"
)

// canHash check file is hashable.
// Hashable file: *.go / llpyg.pub / *.symb.json
func canHash(fileName string) bool {
	if strings.HasSuffix(fileName, ".go") {
		return true
	}
	_, ok := canHashFile[fileName]
	return ok
}

var canHashFile = map[string]struct{}{
	"llpyg.pub": {},
	"go.mod":    {},
	"go.sum":    {},
}

// lockGoVersion locks current Go version to `llpygGoVersion` via GOTOOLCHAIN
func lockGoVersion(cmd *exec.Cmd, pythonPath string) {
	// don't change global settings, use temporary environment.
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOTOOLCHAIN=go%s", llpygGoVersion))
	// Set Python environment if needed
	if pythonPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PYTHONPATH=%s", pythonPath))
	}
}

// llpygGenerator implements Generator interface, which use llpyg tool to generate llpkg.
type llpygGenerator struct {
	dir         string // llpyg.cfg abs path
	pythonDir   string
	packageName string
}

func New(dir, packageName, pythonDir string) generator.Generator {
	return &llpygGenerator{dir: dir, packageName: packageName, pythonDir: pythonDir}
}

// normalizeModulePath returns a normalized module path like
// numpy => github.com/goplus/llpkg/numpy
func (l *llpygGenerator) normalizeModulePath() string {
	return goplusRepo + l.packageName
}

func (l *llpygGenerator) findSymbJSON() string {
	matches, _ := filepath.Glob(filepath.Join(l.dir, "*.symb.json"))
	if len(matches) > 0 {
		return filepath.Base(matches[0])
	}
	return ""
}

func (l *llpygGenerator) copyConfigFileTo(path string) error {
	if l.dir == path {
		return nil
	}
	err := file.CopyFile(
		filepath.Join(l.dir, "llpyg.cfg"),
		filepath.Join(path, "llpyg.cfg"),
	)
	// must stop if llpyg.cfg doesn't exist for safety
	if err != nil {
		return err
	}
	if symb := l.findSymbJSON(); symb != "" {
		file.CopyFile(
			filepath.Join(l.dir, symb),
			filepath.Join(path, symb),
		)
	}
	// ignore copy if file doesn't exist
	file.CopyFile(
		filepath.Join(l.dir, "llpyg.pub"),
		filepath.Join(path, "llpyg.pub"),
	)
	return nil
}

func (l *llpygGenerator) Generate(toDir string) error {
	path, err := filepath.Abs(toDir)
	if err != nil {
		return errors.Join(ErrLLPygGenerate, err)
	}

	// Create a temporary directory for llpyg to work in
	tempWorkDir, err := os.MkdirTemp("", "llpyg-work")
	if err != nil {
		return errors.Join(ErrLLPygGenerate, err)
	}
	defer os.RemoveAll(tempWorkDir)

	// Execute llpyg command directly with package name
	// This is the same as running "llpyg numpy" in the terminal
	cmd := exec.Command("llpyg", l.packageName)
	cmd.Dir = tempWorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// llpyg may exit with an error, which may be caused by Stderr.
	// To avoid that case, we have to check its exit code.
	if err := cmd.Run(); err != nil {
		return errors.Join(ErrLLPygGenerate, err)
	}

	// For Python packages, llpyg generates files in test/packageName directory
	// Check if the generated directory exists
	generatedPath := filepath.Join(tempWorkDir, "test", l.packageName)
	if _, err := os.Stat(generatedPath); os.IsNotExist(err) {
		// Try alternative path
		generatedPath = filepath.Join(tempWorkDir, l.packageName)
		if _, err := os.Stat(generatedPath); os.IsNotExist(err) {
			return errors.Join(ErrLLPygCheck, errors.New("generate fail"))
		}
	}

	// Copy the generated files to the target directory
	// For Python packages, we want to copy the contents of the generated directory
	// to the target directory, not the directory itself
	err = file.CopyFS(path, os.DirFS(generatedPath), true)
	if err != nil {
		return errors.Join(ErrLLPygGenerate, err)
	}

	return nil
}

func (l *llpygGenerator) Check(dir string) error {
	baseDir, err := filepath.Abs(dir)
	if err != nil {
		return errors.Join(ErrLLPygCheck, err)
	}

	// 1. compute hash
	generated, err := hashutils.Dir(baseDir, canHash)
	if err != nil {
		return errors.Join(ErrLLPygCheck, err)
	}
	userGenerated, err := hashutils.Dir(l.dir, canHash)
	if err != nil {
		return errors.Join(ErrLLPygCheck, err)
	}

	// 2. check hash
	for name, hash := range userGenerated {
		generatedHash, ok := generated[name]
		if !ok {
			// if this file is hashable, it's unexpected
			// if not, we can skip it safely.
			if canHash(name) {
				return errors.Join(ErrLLPygCheck, fmt.Errorf("unexpected file: %s", name))
			}
			// skip file
			continue
		}
		if !bytes.Equal(hash, generatedHash) {
			return errors.Join(ErrLLPygCheck, fmt.Errorf("file not equal: %s", name))
		}
	}
	// 3. check missing file
	for name := range generated {
		if _, ok := userGenerated[name]; !ok {
			return errors.Join(ErrLLPygCheck, fmt.Errorf("missing file: %s", name))
		}
	}
	return nil
}
