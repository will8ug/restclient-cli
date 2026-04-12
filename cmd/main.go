package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/will8ug/restclient-cli/internal/parser"
	"github.com/will8ug/restclient-cli/internal/tui"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	arg := os.Args[1]

	if arg == "--version" || arg == "-V" {
		fmt.Printf("restclient %s\n", version)
		os.Exit(0)
	}

	if arg == "--help" || arg == "-h" {
		printUsage()
		os.Exit(0)
	}

	filename := arg

	result, err := parser.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		os.Exit(2)
	}

	if len(result.Requests) == 0 {
		fmt.Fprintf(os.Stderr, "No requests found in %s\n", filename)
		os.Exit(1)
	}

	model := tui.NewModel(result.Requests, filename)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("restclient - TUI REST API client for .http files")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  restclient <file.http>    Launch TUI with requests from file")
	fmt.Println("  restclient --version      Show version")
	fmt.Println("  restclient --help         Show this help")
}
