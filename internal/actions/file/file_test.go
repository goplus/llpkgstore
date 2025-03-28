package file

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestZip(t *testing.T) {
	err := Zip("ziptest", "test.zip")

	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove("test.zip")
	zipr, _ := zip.OpenReader("test.zip")

	exceedMap := map[string]string{
		"ggg.test":                                 "123",
		"ziptest2/gg.test":                         "456",
		"ziptest2/ggg.test":                        "123",
		"ziptest2/ziptest3/gggg.test":              "789",
		"ziptest2/ziptest3/ziptest4/aaa/aaaa.test": "114514",
	}

	compareFile := func(file *zip.File, expect string) {
		fs, err := file.Open()
		if err != nil {
			t.Error(err)
			return
		}
		defer fs.Close()
		b, err := io.ReadAll(fs)
		if err != nil {
			t.Error(err)
			return
		}
		if expect != string(b) {
			t.Errorf("unexpected content: %s: want: %s got: %s", file.Name, expect, string(b))
		}
	}

	fileMap := map[string]struct{}{}
	for _, file := range zipr.File {
		if !file.FileInfo().IsDir() {
			content, ok := exceedMap[file.Name]
			if !ok {
				t.Errorf("unexpected file: %s", file.Name)
			}
			compareFile(file, content)
			fileMap[file.Name] = struct{}{}
		}
	}

	for fileName := range exceedMap {
		if _, ok := fileMap[fileName]; !ok {
			t.Errorf("missing file: %s", fileName)
		}
	}

}

func TestCopyPattern(t *testing.T) {
	os.WriteFile("111.test", []byte("0"), 0644)
	os.WriteFile("222.test", []byte("0"), 0644)

	os.WriteFile("333.test1", []byte("0"), 0644)

	os.Mkdir("aaa", 0777)
	defer os.Remove("111.test")
	defer os.Remove("222.test")
	defer os.Remove("333.test1")
	defer os.RemoveAll("aaa")

	CopyFilePattern(".", "aaa", "*.test")

	fs, _ := os.ReadDir("./aaa")

	expect := map[string]struct{}{
		"111.test": {},
		"222.test": {},
	}

	fileMap := map[string]struct{}{}

	for _, f := range fs {
		if _, ok := expect[f.Name()]; !ok {
			t.Errorf("unexpected file: %s", f.Name())
		}
		fileMap[f.Name()] = struct{}{}
	}

	for fileName := range expect {
		if _, ok := fileMap[fileName]; !ok {
			t.Errorf("missing file: %s", fileName)
		}
	}
}

func TestCopySkip(t *testing.T) {
	from, _ := os.MkdirTemp("", "testcopy-from")
	defer os.RemoveAll(from)

	to, _ := os.MkdirTemp("", "testcopy-to")
	defer os.RemoveAll(to)

	os.WriteFile(filepath.Join(from, "aaa.test"), []byte("123"), 0644)
	os.WriteFile(filepath.Join(to, "aaa.test"), []byte("456"), 0644)

	err := CopyFS(to, os.DirFS(from), true)
	if err != nil {
		t.Error(err)
		return
	}

	toContent, err := os.ReadFile(filepath.Join(to, "aaa.test"))
	if err != nil {
		t.Error(err)
		return
	}

	if string(toContent) != "456" {
		t.Errorf("unexpected skip file: want: 456 got: %s", string(toContent))
	}

	err = CopyFS(to, os.DirFS(from), false)
	if err != nil {
		t.Error(err)
		return
	}
	toContent, err = os.ReadFile(filepath.Join(to, "aaa.test"))
	if err != nil {
		t.Error(err)
		return
	}
	if string(toContent) != "123" {
		t.Errorf("unexpected skip file: want: 123 got: %s", string(toContent))
	}
}
