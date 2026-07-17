package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/loveloki/try/internal/gui"
)

func main() {
	path := flag.String("path", "", "override tries root directory")
	flag.Parse()

	if err := gui.Run(gui.Options{Path: *path}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
