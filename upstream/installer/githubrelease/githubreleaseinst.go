package githubrelease

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	compressPath, err := c.download(c.assertUrl(pkg), outputDir)
	if err != nil {
		return err
	}
	if strings.HasSuffix(compressPath, ".tar.gz") {
		err = c.untargz(outputDir, compressPath)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(compressPath, ".zip") {
		err = c.unzip(outputDir, compressPath)
		if err != nil {
			return err
		}
	}
	err = os.Remove(compressPath)
	if err != nil {
		return fmt.Errorf("cannot delete compressed file: %d", err)
	}
	// err = c.setPrefix(outputDir)
	// if err != nil {
	// 	return fmt.Errorf("fail to reset .pc prefix: %d", err)
	// }
	return nil
}

// Warning: not implemented
func (c *ghReleaseInstaller) Search(pkg upstream.Package) (string, error) {
	// TODO: implement search
	return "", nil
}

func (c *ghReleaseInstaller) assertUrl(pkg upstream.Package) string {
	releaseName := fmt.Sprintf("%s/%s", pkg.Name, pkg.Version)
	fileName := fmt.Sprintf("%s_%s.zip", pkg.Name, c.config["platform"]+"_"+c.config["arch"])
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", c.config["owner"], c.config["repo"], releaseName, fileName)
}

// Download fetches the package from the specified URL and saves it to the output directory.
func (c *ghReleaseInstaller) download(url string, outputDir string) (string, error) {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", ErrPackageNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", err
	}

	outputPath := filepath.Join(outputDir, filename)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// Unzip extracts the gzip-compressed tarball to the output directory.
// The gzipPath must be a .tar.gz file.
func (c *ghReleaseInstaller) untargz(outputDir string, gzipPath string) error {
	fr, err := os.Open(gzipPath)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(outputDir, header.Name)
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}
		dir := filepath.Dir(path)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		if _, err = io.Copy(file, tr); err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

// Unzip extracts the gzip-compressed tarball to the output directory.
// The gzipPath must be a .tar.gz file.
func (c *ghReleaseInstaller) unzip(outputDir string, zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}

	decompress := func(file *zip.File) error {
		path := filepath.Join(outputDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			return nil
		}

		w, err := os.Create(path)
		if err != nil {
			return err
		}
		fs, err := file.Open()
		if err != nil {
			return err
		}
		defer fs.Close()
		io.Copy(w, fs)
		return w.Close()
	}

	for _, file := range r.File {
		if err = decompress(file); err != nil {
			break
		}
	}
	return err
}