package main

import (
	"fmt"
	"io"
)

type terminalWriter struct {
	writer io.Writer
}

func (tw terminalWriter) Printf(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(tw.writer, format, args...); err != nil {
		fmt.Printf("failed to printf: %s", err)
	}
}

func (tw terminalWriter) Print(s string) {
	if _, err := fmt.Fprint(tw.writer, s); err != nil {
		fmt.Printf("failed to print: %s", err)
	}
}

func (tw terminalWriter) Println(s string) {
	tw.Print(s + "\n")
}
