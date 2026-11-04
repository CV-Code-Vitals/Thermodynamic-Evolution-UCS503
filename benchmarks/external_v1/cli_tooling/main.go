package main

import (
	"fmt"
	"os"
)

func main() {
	cfg, err := ParseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("CLITool initialized for cmd=%s, path=%s\n", cfg.Command, cfg.Path)
}

