package client

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func PrintHelp() {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true).
		Render("prx — Reverse Proxy Toolkit")

	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF87")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	text := []string{
		title,
		"",
		descStyle.Render("Usage:") + " prx <command> [flags]",
		"",
		descStyle.Render("Commands:"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("secret"), "Generate a new HMAC secret for JWT signing (Bubble Tea UI)"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("auth"), "Generate a JWT for yourself (Bubble Tea UI)"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("add"), "Add a new redirect via gRPC"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("update"), "Update an existing redirect via gRPC"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("delete"), "Delete a redirect via gRPC"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("list"), "List all redirects via gRPC"),
		fmt.Sprintf("  %s\t%s", cmdStyle.Render("help"), "Show this help"),
		"",
		descStyle.Render("Example:"),
		"  prx secret",
		"  prx auth",
		"  prx add --addr proxy:50051 --token $JWT --from example.com --to http://1.2.3.4 --cert /path/to.crt --key /path/to.key",
	}

	box := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#5555FF")).
		Width(60).
		Render(strings.Join(text, "\n"))

	fmt.Println(box)
}
