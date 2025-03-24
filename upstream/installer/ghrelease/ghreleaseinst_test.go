package ghrelease

import (
	"testing"

	"github.com/goplus/llpkgstore/upstream"
)

func TestGHInstaller(t *testing.T) {
	ghr := &ghReleaseInstaller{
		config: map[string]string{
			"owner": `goplus`,
			"repo":  `llgo`,
		},
	}

	pkg := upstream.Package{
		Version: `v0.10.1`,
		Name:    `llgo0.10.1.darwin-amd64.tar.gz`,
	}

	err := ghr.Install(pkg, `./f`)
	if err != nil {
		t.Errorf("Install failed: %s", err)
	}
}
