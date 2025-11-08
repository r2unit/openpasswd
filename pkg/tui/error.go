package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Error messages with personality
var wrongPassphraseMessages = []string{
	"🔒 Nope! That's not it, chief.",
	"🤔 Nice try, but that's not the magic word.",
	"🙅 Access denied! (Insert dramatic gasp here)",
	"🎯 So close! Just kidding, not even close.",
	"🔐 That passphrase is as wrong as pineapple on pizza.",
	"🚫 Computer says no. Try again, human.",
	"🤖 Error 401: Unauthorized. Also, wrong passphrase.",
	"🎭 Plot twist: That's not your passphrase!",
	"🔑 Wrong key, wrong door, wrong passphrase.",
	"🧙 The passphrase you seek is not the one you speak.",
	"🎪 Congratulations! You've found the wrong passphrase!",
	"🌮 That passphrase is more scrambled than your eggs.",
}

var wrongPassphraseTips = []string{
	"Maybe it's that one from your birthday?",
	"Try the one with the cat's name.",
	"Was it uppercase or lowercase?",
	"Check if Caps Lock is on (classic mistake).",
	"Remember: hunter2 is NOT a secure passphrase.",
	"Did you use the recovery key by mistake?",
	"Coffee first, then passphrase.",
	"Take a deep breath and try again.",
}

type errorModel struct {
	message   string
	tip       string
	width     int
	height    int
	countdown int
}

func NewWrongPassphraseModel() *errorModel {
	rand.Seed(time.Now().UnixNano())
	return &errorModel{
		message:   wrongPassphraseMessages[rand.Intn(len(wrongPassphraseMessages))],
		tip:       wrongPassphraseTips[rand.Intn(len(wrongPassphraseTips))],
		width:     80,
		height:    24,
		countdown: 3,
	}
}

type errorTickMsg time.Time

func (m errorModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return errorTickMsg(t)
	})
}

func (m errorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case errorTickMsg:
		m.countdown--
		if m.countdown <= 0 {
			return m, tea.Quit
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return errorTickMsg(t)
		})

	case tea.KeyMsg:
		// Any key press exits
		return m, tea.Quit
	}

	return m, nil
}

func (m errorModel) View() string {
	var s strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF5F5F")).
		MarginTop(2).
		MarginBottom(1).
		Align(lipgloss.Center)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAF00")).
		MarginBottom(1).
		Align(lipgloss.Center)

	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FAFFF")).
		Italic(true).
		MarginTop(1).
		Align(lipgloss.Center)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginTop(2).
		Align(lipgloss.Center)

	// ASCII art lock
	lockArt := `
    .-"-.
   /  _  \
  |  (_)  |
  |        |
  |  ____  |
  | |____| |
  |________|
  |________|
`

	lockStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F"))

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("⚠️  WRONG PASSPHRASE"))
	s.WriteString("\n\n")
	s.WriteString(lockStyle.Render(lockArt))
	s.WriteString("\n")
	s.WriteString(messageStyle.Render(m.message))
	s.WriteString("\n\n")
	s.WriteString(tipStyle.Render("💡 " + m.tip))
	s.WriteString("\n\n")

	if m.countdown > 0 {
		countdownText := fmt.Sprintf("Closing in %d second%s... (or press any key)",
			m.countdown,
			map[bool]string{true: "", false: "s"}[m.countdown == 1])
		s.WriteString(instructionStyle.Render(countdownText))
	} else {
		s.WriteString(instructionStyle.Render("Press any key to exit..."))
	}

	return s.String()
}

// RunWrongPassphraseTUI shows a fun error screen for wrong passphrase
func RunWrongPassphraseTUI() error {
	model := NewWrongPassphraseModel()
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
