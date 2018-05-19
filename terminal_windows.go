// getColorWriter() for Windows
package main

import (
	"os"

	"github.com/mattn/go-colorable"
)

func getColorWriter(noColor bool) terminalWriter {
	if noColor {
		return terminalWriter{writer: os.Stdout}
	}
	return terminalWriter{writer: colorable.NewColorableStdout()}
}
