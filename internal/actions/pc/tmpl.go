package pc

import (
	"os"
	"path/filepath"
	"regexp"
)

const PCTemplateSuffix = ".tmpl"

var PrefixMatch = regexp.MustCompile(`^prefix=(.*)`)

func GenerateTemplateFromPC(inputName, outputDir string) error {
	pcContent, err := os.ReadFile(inputName)
	if err != nil {
		return err
	}
	outputName := filepath.Join(outputDir, filepath.Base(inputName)+PCTemplateSuffix)
	pcContent = PrefixMatch.ReplaceAll(pcContent, []byte(`prefix={{.Prefix}}`))

	return os.WriteFile(outputName, pcContent, 0644)
}
