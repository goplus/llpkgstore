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

	testPCFile2 = `prefix=/home/vscode/.conan2/p/b/zlibbe9abfe31bec0/p
libdir=${prefix}/lib
includedir=${prefix}/include
includedir1=${prefix}/include/libxml2
bindir=${prefix}/bin

Name: libxml-2.0
Description: Conan package: libxml-2.0
Version: 2.11.6
Libs: -L"${libdir}" -lxml2 -lm -lpthread -ldl
Cflags: -I"${includedir}" -I"${includedir1}"`

	testPCFileCJSON1 = `prefix=/home/vscode/.conan2/p/b/cjson7fb112e50ddea/p
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: cjson
Description: Conan package: cjson
Version: 1.7.18
Libs: -L"${libdir}"
Cflags: -I"${includedir}"
Requires: libcjson libcjson_utils`

	testPCFileCJSON2 = `prefix=/home/vscode/.conan2/p/b/cjson7fb112e50ddea/p
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: cjson
Description: Conan package: cjson
Version: 1.7.18
Libs: -L"${libdir}"
Cflags: -I"${includedir}"
Requires: libcjson libcjson_utils
Requires: zlib`

	testPCFileCJSON3 = `prefix=/home/vscode/.conan2/p/b/cjson7fb112e50ddea/p
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: cjson
Description: Conan package: cjson
Version: 1.7.18
Libs: -L"${libdir}"
Cflags: -I"${includedir}"
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

	expectedContentCJSON = `prefix={{.Prefix}}
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: cjson
Description: Conan package: cjson
Version: 1.7.18
Libs: -L"${libdir}"
Cflags: -I"${includedir}"
Requires: libcjson libcjson_utils`

	expectedContentCJSON2 = `prefix={{.Prefix}}
libdir=${prefix}/lib
includedir=${prefix}/include
bindir=${prefix}/bin

Name: cjson
Description: Conan package: cjson
Version: 1.7.18
Libs: -L"${libdir}"
Cflags: -I"${includedir}"`
)

func TestPCTemplate(t *testing.T) {
	os.WriteFile("test.pc", []byte(testPCFile), 0644)
	os.Mkdir(".generated", 0777)
	GenerateTemplateFromPC("test.pc", ".generated", []string{"libxml-2.0"})
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

func TestPCTemplateNoRequires(t *testing.T) {
	os.WriteFile("test.pc", []byte(testPCFile2), 0644)
	os.Mkdir(".generated", 0777)
	GenerateTemplateFromPC("test.pc", ".generated", []string{"libxml-2.0"})
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

func TestPCTemplateMultiPC(t *testing.T) {
	deps := []string{"cjson", "libcjson", "libcjson_utils"}
	t.Run("internal-require", func(t *testing.T) {
		os.WriteFile("test.pc", []byte(testPCFileCJSON1), 0644)
		os.Mkdir(".generated", 0777)
		GenerateTemplateFromPC("test.pc", ".generated", deps)
		defer os.Remove("test.pc")
		defer os.RemoveAll(".generated")
		b, err := os.ReadFile(filepath.Join(".generated", "test.pc.tmpl"))
		if err != nil {
			t.Error(err)
			return
		}

		if string(b) != expectedContentCJSON {
			t.Errorf("unexpected content: got: %s", string(b))
		}
	})
	t.Run("multi-require", func(t *testing.T) {
		os.WriteFile("test.pc", []byte(testPCFileCJSON2), 0644)
		os.Mkdir(".generated", 0777)
		GenerateTemplateFromPC("test.pc", ".generated", deps)
		defer os.Remove("test.pc")
		defer os.RemoveAll(".generated")
		b, err := os.ReadFile(filepath.Join(".generated", "test.pc.tmpl"))
		if err != nil {
			t.Error(err)
			return
		}

		if string(b) != expectedContentCJSON {
			t.Errorf("unexpected content: got: %s", string(b))
		}
	})

	t.Run("external-require", func(t *testing.T) {
		os.WriteFile("test.pc", []byte(testPCFileCJSON3), 0644)
		os.Mkdir(".generated", 0777)
		GenerateTemplateFromPC("test.pc", ".generated", deps)
		defer os.Remove("test.pc")
		defer os.RemoveAll(".generated")
		b, err := os.ReadFile(filepath.Join(".generated", "test.pc.tmpl"))
		if err != nil {
			t.Error(err)
			return
		}

		if string(b) != expectedContentCJSON2 {
			t.Errorf("unexpected content: got: %s", string(b))
		}
	})
}

func TestABSPathPCTemplate(t *testing.T) {
	pcPath, _ := filepath.Abs("test.pc")
	os.WriteFile(pcPath, []byte(testPCFile), 0644)
	os.Mkdir(".generated", 0777)
	GenerateTemplateFromPC(pcPath, ".generated", []string{"libxml-2.0"})
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
