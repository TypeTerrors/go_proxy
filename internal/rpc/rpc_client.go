package rpc

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func StartClient() {
	if len(os.Args) < 2 {
		PrintHelp()
		os.Exit(0)
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "secret":
		RunSecretUI()
		os.Exit(0)

	case "auth":
		RunAuthUI()
		os.Exit(0)

	case "add", "update", "delete", "list":
		Run(os.Args[1:])
		os.Exit(0)

	case "help":
		PrintHelp()
		os.Exit(0)

	default:
		fmt.Println(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF3333")).
			Bold(true).
			Render("Error:") + " unknown command '" + cmd + "'")
		PrintHelp()
		os.Exit(1)
	}
}
