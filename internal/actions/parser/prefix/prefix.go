package prefix

import (
	"strings"

	"github.com/goplus/llpkgstore/internal/actions/parser"
)

const (
	// LabelPrefix is the prefix identifier for label strings
	LabelPrefix = "branch:"
	// BranchPrefix denotes the prefix format for release branches
	BranchPrefix = "release-branch."
	// MappedVersionPrefix indicates the prefix used for commit version mappings
	MappedVersionPrefix = "Release-as: "
)

// Interface guard to ensure prefixParser implements parser.Parser
var _ parser.Parser = (*prefixParser)(nil)

// prefixParser represents a parser that trims a specific prefix from a string
type prefixParser struct {
	s      string
	prefix string
}

// newPrefixParser creates a new prefixParser instance with the provided string and prefix
func newPrefixParser(s, prefix string) parser.Parser {
	return &prefixParser{s: strings.TrimSpace(s), prefix: prefix}
}

// Parse trims the configured prefix from the input string and returns the result
func (l *prefixParser) Parse() (content string, err error) {
	result := strings.TrimPrefix(l.String(), l.prefix)
	if result == l.String() {
		err = parser.ErrInvalidFormat
		return
	}
	content = result
	return
}

// MustParse parses the string and panics if the prefix is not found
func (l *prefixParser) MustParse() (content string) {
	content, err := l.Parse()
	if err != nil {
		panic(err)
	}
	return
}

// String returns the original input string being parsed
func (l *prefixParser) String() string {
	return l.s
}

// NewLabelParser creates a parser for label strings with the LabelPrefix
func NewLabelParser(content string) parser.Parser {
	return newPrefixParser(content, LabelPrefix)
}

// NewBranchParser creates a parser for branch names with the BranchPrefix
func NewBranchParser(content string) parser.Parser {
	return newPrefixParser(content, BranchPrefix)
}

// NewCommitVersionParser creates a parser for commit version strings with the MappedVersionPrefix
func NewCommitVersionParser(content string) parser.Parser {
	return newPrefixParser(content, MappedVersionPrefix)
}
