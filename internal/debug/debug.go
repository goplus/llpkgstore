package debug

import (
	"os"
	"runtime/debug"
)

func PrintWithStack(err error) {
	os.Stderr.Write([]byte(err.Error() + "\n"))
	debug.PrintStack()
}
