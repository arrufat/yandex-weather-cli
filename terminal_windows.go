package main

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
)

func get_color_writer(no_color bool) io.Writer {
	if no_color {
		return (io.Writer)(os.Stdout)
	}
	return colorable.NewColorableStdout()
}
