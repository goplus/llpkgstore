package pc

import (
	"fmt"
	"os/exec"
)

func SetPath(cmd *exec.Cmd, path string) {
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("PKG_CONFIG_PATH=%s", path))
}
