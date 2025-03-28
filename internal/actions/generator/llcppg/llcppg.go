package llcppg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/internal/actions/generator"
	"github.com/goplus/llpkgstore/internal/actions/hashutils"
	"github.com/goplus/llpkgstore/internal/file"
	"github.com/goplus/llpkgstore/internal/pc"
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
	// default llpkg repo
	goplusRepo = "github.com/goplus/llpkg/"
	// llcppg running default version
	llcppgGoVersion = "1.20.14"
	// llcppg default config file, which MUST exist in specifed dir
	llcppgConfigFile = "llcppg.cfg"
)

// canHash check file is hashable.
// Hashable file: *.go / llcppg.pub / *.symb.json
func canHash(fileName string) bool {
	if strings.HasSuffix(fileName, ".go") {
		return true
	}
	_, ok := canHashFile[fileName]
	return ok
}

// lockGoVersion locks current Go version to `llcppgGoVersion` via GOTOOLCHAIN
func lockGoVersion(cmd *exec.Cmd, pcPath string) {
	// don't change global settings, use temporary environment.
	// see issue: https://github.com/goplus/llpkgstore/issues/18
	pc.SetPath(cmd, pcPath)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOTOOLCHAIN=go%s", llcppgGoVersion))
}

// diffTwoFiles returns the diff result between a file and b file.
func diffTwoFiles(a, b string) string {
	ret, _ := exec.Command("git", "diff", "--no-index", a, b).CombinedOutput()
	return string(ret)
}

func isExitedUnexpectedly(err error) bool {
	process, ok := err.(*exec.ExitError)
	return ok && !process.Success()
}

// llcppgGenerator implements Generator interface, which use llcppg tool to generate llpkg.
type llcppgGenerator struct {
	dir         string // llcppg.cfg abs path
	pcDir       string
	packageName string
}

func New(dir, packageName, pcDir string) generator.Generator {
	return &llcppgGenerator{dir: dir, packageName: packageName, pcDir: pcDir}
}

// normalizeModulePath returns a normalized module path like
// cjson => github.com/goplus/llpkg/cjson
func (l *llcppgGenerator) normalizeModulePath() string {
	return goplusRepo + l.packageName
}

func (l *llcppgGenerator) findSymbJSON() string {
	matches, _ := filepath.Glob(filepath.Join(l.dir, "*.symb.json"))
	if len(matches) > 0 {
		return filepath.Base(matches[0])
	}
	return ""
}

func (l *llcppgGenerator) copyConfigFileTo(path string) error {
	if l.dir == path {
		return nil
	}
	err := file.CopyFile(
		filepath.Join(l.dir, "llcppg.cfg"),
		filepath.Join(path, "llcppg.cfg"),
	)
	// must stop if llcppg.cfg doesn't exist for safety
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
		filepath.Join(l.dir, "llcppg.pub"),
		filepath.Join(path, "llcppg.pub"),
	)
	return nil
}

func (l *llcppgGenerator) Generate(toDir string) error {
	path, err := filepath.Abs(toDir)
	if err != nil {
		return errors.Join(ErrLlcppgGenerate, err)
	}
	if err := l.copyConfigFileTo(path); err != nil {
		return errors.Join(ErrLlcppgGenerate, err)
	}
	cmd := exec.Command("llcppg", "-mod", l.normalizeModulePath(), llcppgConfigFile)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	lockGoVersion(cmd, l.pcDir)

	// llcppg may exit with an error, which may be caused by Stderr.
	// To avoid that case, we have to check its exit code.
	if err := cmd.Run(); isExitedUnexpectedly(err) {
		return errors.Join(ErrLlcppgGenerate, err)
	}
	// check output again
	generatedPath := filepath.Join(path, l.packageName)
	if _, err := os.Stat(generatedPath); os.IsNotExist(err) {
		return errors.Join(ErrLlcppgCheck, errors.New("generate fail"))
	}
	// copy out the generated result
	// be careful: llcppg result MUST not override existed file,
	// otherwise, checking is meaningless.
	err = file.CopyFS(path, os.DirFS(generatedPath), true)
	if err != nil {
		return errors.Join(ErrLlcppgGenerate, err)
	}

	os.RemoveAll(generatedPath)
	return nil
}

func (l *llcppgGenerator) Check(dir string) error {
	baseDir, err := filepath.Abs(dir)
	if err != nil {
		return errors.Join(ErrLlcppgCheck, err)
	}

	// 1. compute hash
	generated, err := hashutils.Dir(baseDir, canHash)
	if err != nil {
		return errors.Join(ErrLlcppgCheck, err)
	}
	userGenerated, err := hashutils.Dir(l.dir, canHash)
	if err != nil {
		return errors.Join(ErrLlcppgCheck, err)
	}

	// 2. check hash
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
			return errors.Join(ErrLlcppgCheck, fmt.Errorf("file not equal: %s %s", name,
				diffTwoFiles(filepath.Join(l.dir, name), filepath.Join(baseDir, name))))
		}
	}
	// 3. check missing file
	for name := range generated {
		if _, ok := userGenerated[name]; !ok {
			return errors.Join(ErrLlcppgCheck, fmt.Errorf("missing file: %s", name))
		}
	}
	return nil
}
