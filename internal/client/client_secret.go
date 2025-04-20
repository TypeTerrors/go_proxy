package client

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	mrand "math/rand" // alias to avoid collisions with crypto/rand
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

const (
	keyLength = 64 // 32 bytes -> 256 bits
)

// generateHMACKey creates a high-entropy secret key for JWT signing.
func generateHMACKey(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// MODEL

type secret struct {
	steps     []string
	index     int
	width     int
	height    int
	spinner   spinner.Model
	progress  progress.Model
	viewFinal bool
	secret    string
	copied    bool
	doneSteps bool
}

// We'll reuse our "installed step" message from the step simulator.
type installedStepMsg string

// STYLES

var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204"))
	currentStepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle        = lipgloss.NewStyle().Margin(1, 2)
	checkMark        = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")

	// Style for the final box that displays the secret.
	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
)

// INIT

func newModel() secret {
	simSteps := []string{
		"Gathering cryptographic entropy",
		"Generating 256-bit HMAC key",
		"Encoding secret",
	}

	// Setup the progress bar
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Setup the spinner
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return secret{
		steps:     simSteps,
		spinner:   s,
		progress:  p,
		viewFinal: false,
	}
}

// COMMANDS & MSG

func (m secret) Init() tea.Cmd {
	// Start the first step
	return tea.Batch(performStep(m.steps[m.index]), m.spinner.Tick)
}

// performStep simulates a short install/generation step.
func performStep(step string) tea.Cmd {
	d := time.Millisecond * time.Duration(mrand.Intn(600)) // fake random delay
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return installedStepMsg(step)
	})
}

// UPDATE

func (m secret) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.viewFinal {
		return finalViewUpdate(m, msg)
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}

	case installedStepMsg:
		step := m.steps[m.index]

		if m.index >= len(m.steps)-1 {
			secret, err := generateHMACKey(keyLength)
			if err != nil {
				log.Fatal("Error generating HMAC key:", err)
			}

			m.doneSteps = true
			m.viewFinal = true
			m.secret = secret

			return m, tea.Printf("%s %s", checkMark, step)
		}

		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.steps)))
		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, step),
			performStep(m.steps[m.index]),
		)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if pm, ok := newModel.(progress.Model); ok {
			m.progress = pm
		}
		return m, cmd
	}

	return m, nil
}

type copyDoneMsg struct {
	err error
}

func copyToClipboardCmd(str string) tea.Cmd {
	return func() tea.Msg {
		clipboard.Write(clipboard.FmtText, []byte(str))
		return copyDoneMsg{nil}
	}
}

// finalViewUpdate handles events once we’re in the “final window”.
func finalViewUpdate(m secret, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit

		case "c":
			return m, copyToClipboardCmd(m.secret)
		}

	case copyDoneMsg:
		if msg.err != nil {
			log.Printf("Error copying to clipboard: %v\n", msg.err)
		} else {
			m.copied = true
		}
		return m, nil
	}

	return m, nil
}

// VIEW

func (m secret) View() string {
	if m.viewFinal {
		return m.finalView()
	}

	return m.stepView()
}

func (m secret) stepView() string {
	if m.doneSteps {
		return doneStyle.Render("Steps complete!\n")
	}

	n := len(m.steps)
	w := lipgloss.Width(fmt.Sprintf("%d", n))
	stepCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()

	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+stepCount))
	stepName := currentStepStyle.Render(m.steps[m.index])
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Running: " + stepName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+stepCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + stepCount
}

// finalView displays the newly generated secret
func (m secret) finalView() string {
	header := titleStyle.Render("🎉 HMAC Key Generated! 🎉")

	secretLine := fmt.Sprintf("Your new HMAC secret is:\n\n  %s", m.secret)
	footer := "(Press c to copy, q to quit)"

	if m.copied {
		footer = "(Copied! Press q to quit)"
	}

	body := fmt.Sprintf("%s\n\n%s\n\n%s", header, secretLine, footer)
	return boxStyle.Render(body)
}

// UTIL

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MAIN

func RunSecretUI() {
	mrand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
