//go:build !windows
// +build !windows

// getColorWriter() for POSIX os-es
package main

import (
	"os"
)

func getColorWriter(_ bool) terminalWriter {
	return terminalWriter{writer: os.Stdout}
}
