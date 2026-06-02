package main

import (
	"os"

	"github.com/alexmchughdev/bootwrangler/internal/cli"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/all"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
