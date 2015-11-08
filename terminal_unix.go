//+build !windows
package main

import (
	"io"
	"os"
)

func get_color_writer(_ bool) io.Writer {
	return (io.Writer)(os.Stdout)
}
