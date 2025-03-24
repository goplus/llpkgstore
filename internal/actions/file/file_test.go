package file

import (
	"archive/zip"
	"io"
	"os"
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
