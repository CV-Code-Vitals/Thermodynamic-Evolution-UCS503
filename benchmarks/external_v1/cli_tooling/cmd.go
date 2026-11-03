package main

import (
	"flag"
	"fmt"
	"io"
)

// Config holds command-line options.
type Config struct {
	Command string
	Path    string
	Verbose bool
}

// ParseFlags parses command line arguments into Config.
func ParseFlags(args []string, output io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("clitool", flag.ContinueOnError)
	fs.SetOutput(output)

	cfg := &Config{}
	fs.StringVar(&cfg.Command, "cmd", "inspect", "command to execute: inspect, scan, report")
	fs.StringVar(&cfg.Path, "path", ".", "target directory path")
	fs.BoolVar(&cfg.Verbose, "v", false, "enable verbose output")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	return cfg, nil
}

