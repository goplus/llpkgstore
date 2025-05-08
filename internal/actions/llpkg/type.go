package llpkg

import "github.com/goplus/llpkgstore/internal/actions/versions"

type (
	PackageName string
	ClibName    string
	ClibVersion string
)

func (p PackageName) String() string {
	return string(p)
}

func (p ClibName) String() string {
	return string(p)
}

func (p ClibVersion) String() string {
	return string(p)
}

func (p ClibVersion) ToSemVer() string {
	return versions.ToSemVer(p.String())
}
