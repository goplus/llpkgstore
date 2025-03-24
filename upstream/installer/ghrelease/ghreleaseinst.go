package ghrelease

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/goplus/llpkgstore/upstream"
)

var ErrPackageNotFound = errors.New("package not found")

// ghReleaseInstaller implements the upstream.Installer interface by downloading
// the corresponding package from a GitHub release.
type ghReleaseInstaller struct {
	config map[string]string
}

// NewGHReleaseInstaller creates a new Conan-based installer instance with provided configuration options.
// The config map supports custom Conan options (e.g., "options": "cjson:utils=True").
func NewGHReleaseInstaller(config map[string]string) upstream.Installer {
	return &ghReleaseInstaller{
		config: config,
	}
}

func (c *ghReleaseInstaller) Name() string {
	return "ghrelease"
}

func (c *ghReleaseInstaller) Config() map[string]string {
	return c.config
}

// Install downloads the package from the GitHub release and extracts it to the output directory.
func (c *ghReleaseInstaller) Install(pkg upstream.Package, outputDir string) error {
	err := c.download(c.assertUrl(pkg), outputDir)
	if err != nil {
		return err
	}
	return nil
}

// Warning: not implemented
func (c *ghReleaseInstaller) Search(pkg upstream.Package) (string, error) {
	// TODO: implement search
	return "", nil
}

func (c *ghReleaseInstaller) assertUrl(pkg upstream.Package) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", c.config["owner"], c.config["repo"], pkg.Version, pkg.Name)
}

func (c *ghReleaseInstaller) download(url string, outputDir string) error {
	client := &http.Client{}
	resp, err := client.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrPackageNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, filename)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
