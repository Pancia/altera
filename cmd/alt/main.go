package main

import (
	"os"

	"github.com/anthropics/altera/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
