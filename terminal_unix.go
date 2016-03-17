// +build !windows

// getColorWriter() for POSIX os-es
package main

import (
	"io"
	"os"
)

func getColorWriter(_ bool) io.Writer {
	return (io.Writer)(os.Stdout)
}
