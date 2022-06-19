package logging

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func IsTTY(out io.Writer) bool {
	if out != os.Stdout && out != os.Stderr {
		return false
	}

	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return false
	} else {
		return true
	}
}
