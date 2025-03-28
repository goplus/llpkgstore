package hashutils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// File hash a file in SHA-256
func File(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil,
			fmt.Errorf("cannot open file: %s err: %v", filePath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil,
			fmt.Errorf("cannot hash file: %s err: %v", filePath, err)
	}
	return h.Sum(nil), nil
}

// Dir hash all the hashable file specified by canHash for the specified directory.
func Dir(dir string, canHash func(string) bool) (fileMap map[string][]byte, err error) {
	fileMap = map[string][]byte{}
	// use ReadDir here instead of filepath.Walk to avoid reading files recursively.
	fs, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, file := range fs {
		// only hashable file
		if !file.IsDir() && canHash(file.Name()) {
			path := filepath.Join(dir, file.Name())
			value, err1 := File(path)
			if err1 != nil {
				err = err1
				break
			}
			fileMap[filepath.Base(path)] = value
		}
	}

	return
}
