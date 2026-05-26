package main

import (
	"os"

	"github.com/xleine/try/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
