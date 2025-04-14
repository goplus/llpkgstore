package parser

import "errors"

var ErrInvalidFormat = errors.New("parse error: invalid format")

type Parser interface {
	Parse() (content string, err error)
	MustParse() (content string)
	String() string
}
