package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/crypto"
)

var (
	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
)

// SetupModel handles the initial setup flow with passphrase and recovery key
type SetupModel struct {
	step              int // 0: welcome, 1: enter passphrase, 2: confirm passphrase, 3: recovery key display
	passphrase        string
	passphraseConfirm string
	recoveryKey       string
	width             int
	height            int
	err               string
	confirmed         bool
	input             string // Current input field
	showPassword      bool
}

type SetupResult struct {
	Passphrase  string
	RecoveryKey string
	Cancelled   bool
}

func NewSetupModel() *SetupModel {
	return &SetupModel{
		step:   0,
		width:  80,
		height: 24,
	}
}

func (m SetupModel) Init() tea.Cmd {
	return nil
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			m.err = ""

		case "tab":
			if m.step == 3 {
				m.showPassword = !m.showPassword
			}

		default:
			// Handle character input for passphrase
			if m.step == 1 || m.step == 2 {
				str := msg.String()
				// Only accept single printable characters (ASCII 32-126)
				if len(str) == 1 && str[0] >= 32 && str[0] <= 126 {
					m.input += str
					m.err = ""
				}
			}
		}
	}

	return m, nil
}

func (m SetupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		// Welcome screen - move to passphrase entry
		m.step = 1
		m.input = ""

	case 1:
		// Validate first passphrase entry
		if len(m.input) < 8 {
			m.err = "Passphrase must be at least 8 characters"
			return m, nil
		}
		m.passphrase = m.input
		m.input = ""
		m.step = 2

	case 2:
		// Validate passphrase confirmation
		if m.input != m.passphrase {
			m.err = "Passphrases do not match"
			m.input = ""
			return m, nil
		}
		m.passphraseConfirm = m.input

		// Generate recovery key
		m.recoveryKey = generateRecoveryKey()
		m.step = 3
		m.input = ""

	case 3:
		// User confirmed they saved recovery key
		if m.confirmed {
			return m, tea.Quit
		}
		// Toggle confirmation
		m.confirmed = true
	}

	return m, nil
}

func (m SetupModel) View() string {
	var s strings.Builder

	switch m.step {
	case 0:
		s.WriteString(m.renderWelcome())
	case 1:
		s.WriteString(m.renderPassphraseEntry())
	case 2:
		s.WriteString(m.renderPassphraseConfirm())
	case 3:
		s.WriteString(m.renderRecoveryKey())
	}

	return s.String()
}

func (m SetupModel) renderWelcome() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("🔐 Welcome to OpenPasswd"))
	s.WriteString("\n\n")
	s.WriteString(infoStyle.Render("Let's set up your secure password manager."))
	s.WriteString("\n\n")

	s.WriteString(normalStyle.Render("You'll need to:"))
	s.WriteString("\n")
	s.WriteString(normalStyle.Render("  1. Create a strong master passphrase"))
	s.WriteString("\n")
	s.WriteString(normalStyle.Render("  2. Save your recovery key (in case you forget your passphrase)"))
	s.WriteString("\n\n")

	s.WriteString(warningStyle.Render("⚠  Important:"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  • Your master passphrase is NEVER stored"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  • We cannot recover your passphrase if you forget it"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  • Keep your recovery key in a safe place"))
	s.WriteString("\n\n")

	s.WriteString(successStyle.Render("→ Press Enter to continue"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  Press Esc to cancel"))
	s.WriteString("\n")

	return s.String()
}

func (m SetupModel) renderPassphraseEntry() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("🔑 Create Master Passphrase"))
	s.WriteString("\n\n")
	s.WriteString(infoStyle.Render("Enter a strong master passphrase:"))
	s.WriteString("\n\n")

	// Password input field
	display := strings.Repeat("•", len(m.input))
	if len(m.input) == 0 {
		display = dimStyle.Render("(min 8 characters)")
	}

	s.WriteString(normalStyle.Render("Passphrase: "))
	s.WriteString(selectedStyle.Render(display))
	s.WriteString("\n\n")

	// Strength indicator
	strength, color := passphraseStrength(m.input)
	if len(m.input) > 0 {
		strengthStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		s.WriteString(strengthStyle.Render(fmt.Sprintf("Strength: %s", strength)))
		s.WriteString("\n\n")
	}

	// Tips
	s.WriteString(dimStyle.Render("Tips for a strong passphrase:"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  • Use 12+ characters"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  • Mix uppercase, lowercase, numbers, symbols"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("  • Use a phrase you'll remember but others can't guess"))
	s.WriteString("\n\n")

	if m.err != "" {
		s.WriteString(errorStyle.Render("✗ " + m.err))
		s.WriteString("\n\n")
	}

	s.WriteString(dimStyle.Render("Press Enter to continue • Esc to cancel"))
	s.WriteString("\n")

	return s.String()
}

func (m SetupModel) renderPassphraseConfirm() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("🔑 Confirm Master Passphrase"))
	s.WriteString("\n\n")
	s.WriteString(infoStyle.Render("Enter your passphrase again to confirm:"))
	s.WriteString("\n\n")

	display := strings.Repeat("•", len(m.input))
	if len(m.input) == 0 {
		display = dimStyle.Render("(confirm passphrase)")
	}

	s.WriteString(normalStyle.Render("Confirm: "))
	s.WriteString(selectedStyle.Render(display))
	s.WriteString("\n\n")

	if m.err != "" {
		s.WriteString(errorStyle.Render("✗ " + m.err))
		s.WriteString("\n\n")
	}

	s.WriteString(dimStyle.Render("Press Enter to continue • Esc to cancel"))
	s.WriteString("\n")

	return s.String()
}

func (m SetupModel) renderRecoveryKey() string {
	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(titleStyle.Render("🔐 Your Recovery Key"))
	s.WriteString("\n\n")
	s.WriteString(warningStyle.Render("⚠  IMPORTANT: Save this recovery key!"))
	s.WriteString("\n\n")
	s.WriteString(infoStyle.Render("If you forget your master passphrase, this recovery key"))
	s.WriteString("\n")
	s.WriteString(infoStyle.Render("is the ONLY way to access your passwords."))
	s.WriteString("\n\n")

	// Recovery key display (formatted nicely)
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#1A1A1A")).
		Bold(true).
		Padding(1, 2)

	s.WriteString(keyStyle.Render(formatRecoveryKey(m.recoveryKey)))
	s.WriteString("\n\n")

	s.WriteString(dimStyle.Render("📝 Write this down or save it in a password manager"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("🔒 Store it somewhere safe (not on this computer)"))
	s.WriteString("\n")
	s.WriteString(dimStyle.Render("⚠  Never share this key with anyone"))
	s.WriteString("\n\n")

	if !m.confirmed {
		s.WriteString(successStyle.Render("☐ I have saved my recovery key"))
		s.WriteString("\n\n")
		s.WriteString(dimStyle.Render("Press Enter to confirm • Tab to toggle"))
	} else {
		s.WriteString(successStyle.Render("☑ I have saved my recovery key"))
		s.WriteString("\n\n")
		s.WriteString(successStyle.Render("→ Press Enter to finish setup"))
	}
	s.WriteString("\n")

	return s.String()
}

// passphraseStrength calculates passphrase strength
func passphraseStrength(pass string) (string, string) {
	if len(pass) < 8 {
		return "Too Short", "#FF5F5F"
	}

	score := 0
	if len(pass) >= 12 {
		score++
	}
	if len(pass) >= 16 {
		score++
	}

	hasLower := strings.ContainsAny(pass, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(pass, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(pass, "0123456789")
	hasSpecial := strings.ContainsAny(pass, "!@#$%^&*()_+-=[]{}|;:,.<>?")

	if hasLower {
		score++
	}
	if hasUpper {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	switch {
	case score <= 2:
		return "Weak", "#FF5F5F"
	case score <= 4:
		return "Fair", "#FFAF00"
	case score <= 5:
		return "Good", "#5FAFFF"
	default:
		return "Strong", "#00FF00"
	}
}

// generateRecoveryKey generates a secure recovery key
func generateRecoveryKey() string {
	key, err := crypto.GenerateRecoveryKeyWithChecksum()
	if err != nil {
		// Fallback to simple key if generation fails
		return "error-generating-recovery-key"
	}
	return key
}

// formatRecoveryKey formats the recovery key for display
func formatRecoveryKey(key string) string {
	return crypto.FormatRecoveryKey(key)
}

// RunSetupTUI runs the initial setup flow
func RunSetupTUI() (SetupResult, error) {
	model := NewSetupModel()
	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return SetupResult{Cancelled: true}, err
	}

	finalModel := result.(SetupModel)

	if !finalModel.confirmed {
		return SetupResult{Cancelled: true}, nil
	}

	return SetupResult{
		Passphrase:  finalModel.passphrase,
		RecoveryKey: finalModel.recoveryKey,
		Cancelled:   false,
	}, nil
}
