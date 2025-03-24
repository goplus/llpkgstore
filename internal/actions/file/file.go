package file

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyFS copies the file system fsys into the directory dir,
// creating dir if necessary.
//
// Files are created with mode 0o666 plus any execute permissions
// from the source, and directories are created with mode 0o777
// (before umask).
//
// CopyFS will not overwrite existing files. If a file name in fsys
// already exists in the destination, CopyFS will return an error
// such that errors.Is(err, fs.ErrExist) will be true.
//
// Symbolic links in fsys are not supported. A *PathError with Err set
// to ErrInvalid is returned when copying from a symbolic link.
//
// Symbolic links in dir are followed.
//
// Copying stops at and returns the first error encountered.
func CopyFS(dir string, fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		newPath := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(newPath, 0777)
		}

		// TODO(panjf2000): handle symlinks with the help of fs.ReadLinkFS
		// 		once https://go.dev/issue/49580 is done.
		//		we also need filepathlite.IsLocal from https://go.dev/cl/564295.
		if !d.Type().IsRegular() {
			return &os.PathError{Op: "CopyFS", Path: path, Err: os.ErrInvalid}
		}

		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		info, err := r.Stat()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(newPath, os.O_TRUNC|os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666|info.Mode()&0777)
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, r); err != nil {
			w.Close()
			return &os.PathError{Op: "Copy", Path: newPath, Err: err}
		}
		return w.Close()
	})
}

// CopyFile copies a file from the source path 'from' to the destination path 'to'.
//
// It opens the source file, creates the destination file (overwriting if exists),
// and copies the contents. The destination file permissions are determined by
// the os.Create default mode modified by any umask.
//
// Returns an error if opening the source, creating the destination, or copying
// the contents fails.
func CopyFile(from, to string) (err error) {
	r, err := os.Open(from)
	if err != nil {
		return
	}
	defer r.Close()
	w, err := os.Create(to)
	if err != nil {
		return
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return
}

// Zip zips a directory.
func Zip(zipDir, fileName string) error {
	zipFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	return filepath.WalkDir(zipDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		normalizePath, err := filepath.Rel(zipDir, path)
		if err != nil {
			return err
		}
		if normalizePath == "." {
			return nil
		}
		if d.IsDir() {
			normalizePath = fmt.Sprintf("%s%c", normalizePath, os.PathSeparator)
		}
		zw, err := w.Create(normalizePath)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		fs, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(zw, fs)
		if err != nil {
			return err
		}
		return fs.Close()
	})
}

func CopyFilePattern(from, to, pattern string) (err error) {
	matches, err := filepath.Glob(filepath.Join(from, pattern))
	if err != nil {
		return err
	}
	for _, match := range matches {
		err = CopyFile(match, filepath.Join(to, filepath.Base(match)))
		if err != nil {
			break
		}
	}
	return
}
