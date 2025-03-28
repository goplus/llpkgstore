package githubrelease

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goplus/llpkgstore/upstream"
)

var ErrPackageNotFound = errors.New("package not found")

// ghReleaseInstaller implements the upstream.Installer interface by downloading
// the corresponding package from a GitHub release.
type ghReleaseInstaller struct {
	config map[string]string
}

// NewGHReleaseInstaller creates a new GitHub Release installer with the specified configuration.
// The config map is the info of release repo, for example:
// "owner":    `goplus`,
// "repo":     `llpkg`,
// "platform": runtime.GOOS,
// "arch":     runtime.GOARCH,
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

// Install downloads the package from the GitHub Release and extracts it to the output directory.
// Unlike conaninstaller which is used for GitHub Action to obtain binary files
// this installer is used for `llgo get` to install binary files.
// The first return value is an empty string, as the pkgConfigName is not necessary for this GitHub Release installer.
func (c *ghReleaseInstaller) Install(pkg upstream.Package, outputDir string) (string, error) {
	compressPath, err := c.download(c.assertUrl(pkg), outputDir)
	if err != nil {
		return "", err
	}
	if strings.HasSuffix(compressPath, ".tar.gz") {
		err = c.untargz(outputDir, compressPath)
		if err != nil {
			return "", err
		}
	} else if strings.HasSuffix(compressPath, ".zip") {
		err = c.unzip(outputDir, compressPath)
		if err != nil {
			return "", err
		}
	}
	err = os.Remove(compressPath)
	if err != nil {
		return "", errors.Join(errors.New("cannot delete compressed file: "), err)
	}
	err = c.setPrefix(outputDir)
	if err != nil {
		return "", errors.Join(errors.New("fail to reset .pc prefix: "), err)
	}
	return "", nil
}

// Warning: not implemented
// Search is unnecessary for this installer
func (c *ghReleaseInstaller) Search(pkg upstream.Package) ([]string, error) {
	return nil, nil
}

// assertUrl returns the URL for the specified package.
// The URL is constructed based on the package name, version, and the installer configuration.
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
// The gzipPath must be a .zip file.
func (c *ghReleaseInstaller) unzip(outputDir string, zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}

	decompress := func(file *zip.File) error {
		path := filepath.Join(outputDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
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
		if _, err := io.Copy(w, fs); err != nil {
			return err
		}
		return w.Close()
	}

	for _, file := range r.File {
		if err = decompress(file); err != nil {
			break
		}
	}
	return err
}

// Generate .pc files from .pc.tmpl files
func (c *ghReleaseInstaller) setPrefix(outputDir string) error {
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}

	// move to path where .pc files are stored
	pkgConfigPath := filepath.Join(outputDir, "lib/pkgconfig")

	pcTmpls, err := filepath.Glob(filepath.Join(pkgConfigPath, "*.pc.tmpl"))
	if err != nil {
		return err
	}
	if len(pcTmpls) == 0 {
		return errors.New("no .pc.tmpl files found")
	}
	for _, pcTmpl := range pcTmpls {
		tmplContent, err := os.ReadFile(pcTmpl)
		if err != nil {
			return err
		}
		tmplName := filepath.Base(pcTmpl)
		tmpl, err := template.New(tmplName).Parse(string(tmplContent))
		if err != nil {
			return err
		}
		// The Prefix field specifies the absolute path to the output directory,
		// which is used to replace placeholders in the .pc template files.
		data := struct {
			Prefix string
		}{
			Prefix: absOutputDir,
		}
		pcFilePath := filepath.Join(pkgConfigPath, strings.TrimSuffix(tmplName, ".tmpl"))

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return err
		}
		if err := os.WriteFile(pcFilePath, buf.Bytes(), 0644); err != nil {
			return err
		}
		// remove .pc.tmpl file
		err = os.Remove(filepath.Join(pkgConfigPath, tmplName))
		if err != nil {
			return errors.Join(errors.New("failed to remove template file: "), err)
		}
	}

	return nil
}
