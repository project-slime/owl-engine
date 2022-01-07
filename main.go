package main

import (
	"os"

	"owl-engine/cmd/engine"
)

func main() {
	if err := engine.Execute(); err != nil {
		os.Exit(1)
	}
}
