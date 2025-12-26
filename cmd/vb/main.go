package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dragonbytelabs/voidabyss/internal/editor"
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
		// Treat as file path
		runEditor(args[0])
	}
}

func runEditor(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(abs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid path: %v\n", err)
		os.Exit(1)
	}

	var runErr error
	if info.IsDir() {
		runErr = editor.OpenProject(abs)
	} else {
		runErr = editor.OpenFile(abs)
	}

	if runErr != nil {
    		fmt.Fprintf(os.Stderr, "voidabyss error: %v\n", runErr)
		os.Exit(2)
	}

}

func printHelp() {
	fmt.Print(`
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
