package rpc

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.design/x/clipboard"

	"prx/internal/services"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// -----------------------------------------------------------------------------
// Color Palette
// -----------------------------------------------------------------------------

var (
	colorPrimary   = lipgloss.Color("#04B575") // Teal-like green
	colorSecondary = lipgloss.Color("#FFA500") // Orange
	colorError     = lipgloss.Color("#FF3333") // Bright red
	colorText      = lipgloss.Color("#CCCCCC") // Grayish
	colorAccent    = lipgloss.Color("#5555FF") // A nice blue
)

// -----------------------------------------------------------------------------
// Lip Gloss Styles
// -----------------------------------------------------------------------------

var baseStyle = lipgloss.NewStyle().
	Foreground(colorText)

var headingStyle = lipgloss.NewStyle().
	Foreground(colorSecondary).
	Bold(true)

var errorStyle = lipgloss.NewStyle().
	Foreground(colorError).
	Bold(true)

var subtleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#999999"))

var tokenBorderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colorAccent).
	Padding(0, 1)

// -----------------------------------------------------------------------------
// JWT “Generation” (Simulated)
// -----------------------------------------------------------------------------

func generateJWT(secret string) (string, error) {

	jwtSvc := services.NewJwtService(secret)

	// 4) Generate a new token
	tokenStr, err := jwtSvc.GenerateJWT()
	if err != nil {
		log.Fatal("Failed to generate JWT", "err", err)
	}
	return tokenStr, nil
}

// -----------------------------------------------------------------------------
// States
// -----------------------------------------------------------------------------

type state int

const (
	stateInputSecret state = iota
	stateGenerating
	stateResult
)

// -----------------------------------------------------------------------------
// Model
// -----------------------------------------------------------------------------

type auth struct {
	state state

	// Bubbles
	input    textinput.Model
	progress progress.Model

	// Progress percentage
	percent float64

	// Data
	secret   string
	jwtToken string
	copied   bool
	err      error
}

// -----------------------------------------------------------------------------
// Messages
// -----------------------------------------------------------------------------

type tickMsg time.Time

// -----------------------------------------------------------------------------
// Commands
// -----------------------------------------------------------------------------

// We’ll tick every half second to advance the progress bar
func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// -----------------------------------------------------------------------------
// Init
// -----------------------------------------------------------------------------

func (m auth) Init() tea.Cmd {
	// Start by blinking the text input cursor
	return textinput.Blink
}

// -----------------------------------------------------------------------------
// Update
// -----------------------------------------------------------------------------

func (m auth) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {

	// 1) Input phase
	case stateInputSecret:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				m.secret = m.input.Value()
				// Move to generating state, start progress at 0
				m.state = stateGenerating
				m.percent = 0
				return m, tickCmd()
			}
		}
		// Let the text input handle other messages
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	// 2) Generating (progress bar)
	case stateGenerating:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Allow quitting with ctrl+c
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}

		case tickMsg:
			// Advance progress by 25% every half second
			m.percent += 0.25
			if m.percent >= 1.0 {
				// Done. Generate the JWT, move to result
				token, err := generateJWT(m.secret)
				if err != nil {
					m.err = err
				} else {
					m.jwtToken = token
				}
				m.state = stateResult
				return m, nil
			}
			return m, tickCmd()
		}
		return m, nil

	// 3) Result
	case stateResult:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "c":
				// Copy the JWT
				return m, copyToClipboardCmd(m.jwtToken)
			}
		case copyDoneMsg:
			if msg.err == nil {
				m.copied = true
			} else {
				m.err = msg.err
			}
			return m, nil
		}
	}

	return m, nil
}

// -----------------------------------------------------------------------------
// View
// -----------------------------------------------------------------------------

func (m auth) View() string {
	switch m.state {
	// 1) Input Secret
	case stateInputSecret:
		return baseStyle.Render(fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			headingStyle.Render("Enter your JWT Secret:"),
			m.input.View(),
			subtleStyle.Render("Running version "+Version+". Press Enter to generate. Press Ctrl+C to quit."),
		))

	// 2) Generating (progress bar)
	case stateGenerating:
		progressBar := m.progress.ViewAs(m.percent)
		pad := strings.Repeat(" ", 2)
		return baseStyle.Render(fmt.Sprintf(
			"\n%s\n\n%s%s\n\n%s\n",
			headingStyle.Render("Generating your token..."),
			pad,
			progressBar,
			subtleStyle.Render("Wait until 100% or press Ctrl+C to quit."),
		))

	// 3) Show final token
	case stateResult:
		// If there’s an error
		if m.err != nil {
			return errorStyle.Render(fmt.Sprintf(
				"\nError generating JWT: %v\n\nPress Ctrl+C to exit.\n", m.err,
			))
		}

		copyHint := "Press 'c' to copy the token, or Ctrl+C to quit."
		if m.copied {
			copyHint = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				Render("✔ Copied to clipboard!")
		}

		tokenBox := tokenBorderStyle.Render(m.jwtToken)

		return baseStyle.Render(fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			headingStyle.Render("Your JWT is:"),
			tokenBox,
			subtleStyle.Render(copyHint),
		))
	}

	return "Unknown state."
}

// -----------------------------------------------------------------------------
// Main
// -----------------------------------------------------------------------------
var Version = "N/A"

func RunAuthUI() {
	// Initialize the clipboard library
	if err := clipboard.Init(); err != nil {
		fmt.Println("Clipboard init error:", err)
		os.Exit(1)
	}

	// Configure the text input
	ti := textinput.New()
	ti.Placeholder = "Secret"
	ti.Focus()

	// Configure the progress bar
	prog := progress.New(
		progress.WithScaledGradient(string(colorSecondary), string(colorPrimary)),
		progress.WithWidth(30),
	)

	// Create our initial model
	m := auth{
		state:    stateInputSecret,
		input:    ti,
		progress: prog,
		percent:  0,
	}

	// Run the program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
