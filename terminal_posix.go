// +build !windows

// getColorWriter() for POSIX os-es
package main

import (
	"io"
	"os"
)

func getColorWriter(_ bool) terminalWriter {
	return terminalWriter{writer: (io.Writer)(os.Stdout)}
}
