package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func errPrintf(format string, args... interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	flag.Parse()
	filenames := flag.Args()
	fset := token.NewFileSet()
	_, firstErr := parser.ParseFiles(fset, filenames, 0)
	if firstErr != nil {
		errPrintf("Error while parsing: %v", firstErr)
	}
}
