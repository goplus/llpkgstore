package pc

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testPCFile = `prefix=/home/vscode/.conan2/p/b/zlibbe9abfe31bec0/p
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: zlib
Description: Conan package: zlib
Version: 1.3.1
Libs: -L"${libdir}" -lz
Cflags: -I"${includedir}"`

	expectedContent = `prefix={{.Prefix}}
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: zlib
Description: Conan package: zlib
Version: 1.3.1
Libs: -L"${libdir}" -lz
Cflags: -I"${includedir}"`
)

func TestPCTemplate(t *testing.T) {
	os.WriteFile("test.pc", []byte(testPCFile), 0644)
	os.Mkdir(".generated", 0777)
	GenerateTemplateFromPC("test.pc", ".generated")
	defer os.Remove("test.pc")
	defer os.RemoveAll(".generated")
	b, err := os.ReadFile(filepath.Join(".generated", "test.pc.tmpl"))
	if err != nil {
		t.Error(err)
		return
	}

	if string(b) != expectedContent {
		t.Errorf("unexpected content: got: %s", string(b))
	}
}

func TestABSPathPCTemplate(t *testing.T) {
	pcPath, _ := filepath.Abs("test.pc")
	os.WriteFile(pcPath, []byte(testPCFile), 0644)
	os.Mkdir(".generated", 0777)
	GenerateTemplateFromPC(pcPath, ".generated")
	defer os.Remove(pcPath)
	defer os.RemoveAll(".generated")
	b, err := os.ReadFile(filepath.Join(".generated", "test.pc.tmpl"))
	if err != nil {
		t.Error(err)
		return
	}

	if string(b) != expectedContent {
		t.Errorf("unexpected content: got: %s", string(b))
	}
}
