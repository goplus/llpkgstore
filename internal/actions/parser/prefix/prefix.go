package prefix

import (
	"strings"

	"github.com/goplus/llpkgstore/internal/actions/parser"
)

const (
	LabelPrefix         = "branch:"
	BranchPrefix        = "release-branch."
	MappedVersionPrefix = "Release-as: "
)

var _ parser.Parser = (*prefixParser)(nil)

type prefixParser struct {
	s      string
	prefix string
}

func newPrefixParser(s, prefix string) parser.Parser {
	return &prefixParser{s: strings.TrimSpace(s), prefix: prefix}
}

func (l *prefixParser) Parse() (content string, err error) {
	result := strings.TrimPrefix(l.String(), l.prefix)
	if result == l.String() {
		err = parser.ErrInvalidFormat
		return
	}
	content = result
	return
}

func (l *prefixParser) MustParse() (content string) {
	content, err := l.Parse()
	if err != nil {
		panic(err)
	}
	return
}

func (l *prefixParser) String() string {
	return l.s
}

func NewLabelParser(content string) parser.Parser {
	return newPrefixParser(content, LabelPrefix)
}

func NewBranchParser(content string) parser.Parser {
	return newPrefixParser(content, BranchPrefix)
}

func NewCommitVersionParser(content string) parser.Parser {
	return newPrefixParser(content, MappedVersionPrefix)
}
