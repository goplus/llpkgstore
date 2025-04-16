package env

import "fmt"

type envError struct {
	Name string
}

func newEnvError(name string) error {
	return &envError{Name: name}
}

func (e *envError) Error() string {
	return fmt.Sprintf("env: no %s found", e.Name)
}
