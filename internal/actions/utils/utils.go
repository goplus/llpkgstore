package utils

import (
	"os"
	"os/exec"
)

func OutputToStdout(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
