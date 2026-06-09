package main

import (
	"os"

	"github.com/meshbrow-dev/meshbrow-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
