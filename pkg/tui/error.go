package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/r2unit/openpasswd/pkg/config"
)

// Default error messages with personality
var defaultWrongPassphraseMessages = []string{
	"Nope! That's not it, chief.",
	"Nice try, but that's not the magic word.",
	"Access denied! (Insert dramatic gasp here)",
	"So close! Just kidding, not even close.",
	"That passphrase is as wrong as pineapple on pizza.",
	"Computer says no. Try again, human.",
	"Error 401: Unauthorized. Also, wrong passphrase.",
	"Plot twist: That's not your passphrase!",
	"Wrong key, wrong door, wrong passphrase.",
	"The passphrase you seek is not the one you speak.",
	"Congratulations! You've found the wrong passphrase!",
	"That passphrase is more scrambled than your eggs.",
}

var defaultWrongPassphraseTips = []string{
	"Maybe it's that one from your birthday?",
	"Was it uppercase or lowercase?",
	"Check if Caps Lock is on (classic mistake).",
	"Remember: hunter2 is NOT a secure passphrase.",
	"Did you use the recovery key by mistake?",
	"Coffee first, then passphrase.",
	"Take a deep breath and try again.",
}

type errorModel struct {
	message     string
	tip         string
	width       int
	height      int
	countdown   int
	shouldRetry bool
}

func NewWrongPassphraseModel() *errorModel {
	rand.Seed(time.Now().UnixNano())

	// Load custom messages from config, or use defaults
	messages := defaultWrongPassphraseMessages
	tips := defaultWrongPassphraseTips

	if customMessages := config.LoadErrorMessages(); len(customMessages) > 0 {
		messages = customMessages
	}
	if customTips := config.LoadErrorTips(); len(customTips) > 0 {
		tips = customTips
	}

	return &errorModel{
		message:     messages[rand.Intn(len(messages))],
		tip:         tips[rand.Intn(len(tips))],
		width:       80,
		height:      24,
		countdown:   3,
		shouldRetry: false,
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
		key := msg.String()
		// ESC or ctrl+c quits without retry
		if key == "esc" || key == "ctrl+c" {
			m.shouldRetry = false
			return m, tea.Quit
		}
		// Any other key triggers retry
		m.shouldRetry = true
		return m, tea.Quit
	}

	return m, nil
}

func (m errorModel) View() string {
	var s strings.Builder

	// Tree-style header
	s.WriteString(addNormalStyle.Render("┌  "))
	s.WriteString(addErrorStyle.Render("Wrong passphrase"))
	s.WriteString("\n")
	s.WriteString(addNormalStyle.Render("│"))
	s.WriteString("\n")

	// Error message
	s.WriteString(addNormalStyle.Render("│  "))
	s.WriteString(addSelectedStyle.Render("✗  "))
	s.WriteString(addSelectedStyle.Render(m.message))
	s.WriteString("\n")
	s.WriteString(addNormalStyle.Render("│"))
	s.WriteString("\n")

	// Helpful tip
	s.WriteString(addNormalStyle.Render("│  "))
	s.WriteString(listMetaStyle.Render("💡  "))
	s.WriteString(listMetaStyle.Render(m.tip))
	s.WriteString("\n")
	s.WriteString(addNormalStyle.Render("│"))
	s.WriteString("\n")

	// Footer with countdown
	s.WriteString(addNormalStyle.Render("└  "))
	if m.countdown > 0 {
		countdownText := fmt.Sprintf("Press any key to retry • Closing in %d second%s",
			m.countdown,
			map[bool]string{true: "", false: "s"}[m.countdown == 1])
		s.WriteString(listMetaStyle.Render(countdownText))
	} else {
		s.WriteString(listMetaStyle.Render("Press any key to retry • esc to quit"))
	}

	return s.String()
}

// RunWrongPassphraseTUI shows a fun error screen for wrong passphrase
// Returns (shouldRetry bool, error)
func RunWrongPassphraseTUI() (bool, error) {
	model := NewWrongPassphraseModel()
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	// Extract the final model state
	if m, ok := finalModel.(errorModel); ok {
		return m.shouldRetry, nil
	}

	return false, nil
}
