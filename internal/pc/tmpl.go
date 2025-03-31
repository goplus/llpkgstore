package pc

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const PCTemplateSuffix = ".tmpl"

var (
	// By the way, trim the new line
	requireMatch = regexp.MustCompile(`\nRequires:\s(.*)`)
	PrefixMatch  = regexp.MustCompile(`^prefix=(.*)`)
)

func isInternalDeps(s []string, internalDeps []string) bool {
	// TODO(ghl): optimize this function, O(m*n) is slow.
	m := make(map[string]struct{}, len(s))

	for _, str := range s {
		m[str] = struct{}{}
	}

	for _, str := range internalDeps {
		if _, ok := m[str]; ok {
			return true
		}
	}
	return false
}

func GenerateTemplateFromPC(inputName, outputDir string, internalDeps []string) error {
	pcContent, err := os.ReadFile(inputName)
	if err != nil {
		return err
	}
	outputName := filepath.Join(outputDir, filepath.Base(inputName)+PCTemplateSuffix)
	pcContent = PrefixMatch.ReplaceAll(pcContent, []byte(`prefix={{.Prefix}}`))

	for _, ret := range requireMatch.FindAllSubmatch(pcContent, -1) {
		// check it's an external deps or not
		requireName := strings.Fields(string(ret[1]))

		if !isInternalDeps(requireName, internalDeps) {
			// it's an external deps, can remove.
			pcContent = bytes.ReplaceAll(pcContent, ret[0], []byte(""))
		}
	}

	return os.WriteFile(outputName, pcContent, 0644)
}
