// getColorWriter() for Windows
package main

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
)

func getColorWriter(noColor bool) terminalWriter {
	if noColor {
		return terminalWriter{writer: (io.Writer)(os.Stdout)}
	}
	return terminalWriter{writer: colorable.NewColorableStdout()}
}
