package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dragonbytelabs/voidabyss/internal/app"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// default: open current directory
		runEditor(".")
		return
	}

	switch args[0] {
	case "edit":
		path := "."
		if len(args) > 1 {
			path = args[1]
		}
		runEditor(path)

	case "version", "--version", "-v":
		fmt.Println("VoidAbyss", version)

	case "help", "--help", "-h":
		printHelp()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func runEditor(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid path: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(abs); err != nil {
		fmt.Fprintf(os.Stderr, "voidabyss error: %v\n", err)
		os.Exit(2)
	}
}

func printHelp() {
	fmt.Println(`
VoidAbyss — a minimal, modal environment for deep work.

Usage:
  vb
  vb edit [path]
  vb version
  vb help

Examples:
  vb
  vb edit .
  vb edit main.go
`)
}