// getColorWriter() for Windows
package main

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
)

func getColorWriter(noColor bool) io.Writer {
	if noColor {
		return (io.Writer)(os.Stdout)
	}
	return colorable.NewColorableStdout()
}
