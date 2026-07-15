package main

import (
	"os"

	"github.com/B4Dmonkey/bit-pro/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
