package main

import (
	"os"

	"github.com/pratik-mahalle/infraudit/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
