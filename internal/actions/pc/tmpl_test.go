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
includedir1=${prefix}/include/libxml2
bindir=${prefix}/bin

Name: libxml-2.0
Description: Conan package: libxml-2.0
Version: 2.11.6
Libs: -L"${libdir}" -lxml2 -lm -lpthread -ldl
Cflags: -I"${includedir}" -I"${includedir1}"
Requires: zlib`

	expectedContent = `prefix={{.Prefix}}
libdir=${prefix}/lib
includedir=${prefix}/include
includedir1=${prefix}/include/libxml2
bindir=${prefix}/bin

Name: libxml-2.0
Description: Conan package: libxml-2.0
Version: 2.11.6
Libs: -L"${libdir}" -lxml2 -lm -lpthread -ldl
Cflags: -I"${includedir}" -I"${includedir1}"`
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
