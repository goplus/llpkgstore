package actions

import (
	"errors"
	"fmt"
)

var (
	ErrNoMappedVersion = errors.New("actions: no MappedVersion found")
)

type actionError struct {
	Err error
}

func wrapActionError(err error) error {
	if err == nil {
		return nil
	}
	return &actionError{Err: err}
}

func (a *actionError) Error() string {
	return fmt.Sprintf("actions: %v", a.Err)
}
