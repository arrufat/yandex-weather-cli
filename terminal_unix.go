// +build !windows

package main

import (
	"io"
	"os"
)

func getColorWriter(_ bool) io.Writer {
	return (io.Writer)(os.Stdout)
}
