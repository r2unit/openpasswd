package tui

// init.go provides a TUI interface for prompting users when configuration already exists.
// It offers two choices: Ignore (keep existing) or Override (replace with new).
// When Override is selected, a confirmation prompt is shown to prevent accidental data loss.

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type initModel struct {
	cursor   int
	selected int
	step     int // 0: choose action, 1: confirm override
	width    int
	height   int
}

var (
	initTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#5FAFFF")).
			MarginBottom(1)

	initWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	initDangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			Bold(true)

	initSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	initNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8"))

	initInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5FAFFF"))
)

type InitChoice int

const (
	ChoiceIgnore InitChoice = iota
	ChoiceOverride
	ChoiceCancel
)

func NewInitTUI() *initModel {
	return &initModel{
		cursor:   0,
		selected: -1,
		step:     0,
		width:    80,
		height:   24,
	}
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.selected = int(ChoiceCancel)
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxCursor := 1
			if m.step == 1 {
				maxCursor = 1 // Yes/No for confirmation
			}
			if m.cursor < maxCursor {
				m.cursor++
			}

		case "enter":
			if m.step == 0 {
				if m.cursor == 0 {
					// Ignore - keep existing
					m.selected = int(ChoiceIgnore)
					return m, tea.Quit
				} else if m.cursor == 1 {
					// Override - go to confirmation
					m.step = 1
					m.cursor = 1 // Default to "No"
				}
			} else if m.step == 1 {
				if m.cursor == 0 {
					// Yes - confirm override
					m.selected = int(ChoiceOverride)
					return m, tea.Quit
				} else {
					// No - cancel
					m.selected = int(ChoiceCancel)
					return m, tea.Quit
				}
			}
		}
	}

	return m, nil
}

func (m initModel) View() string {
	var s strings.Builder

	if m.step == 0 {
		// Step 1: Choose action
		s.WriteString("\n")
		s.WriteString(initTitleStyle.Render("⚠  Configuration Already Exists"))
		s.WriteString("\n\n")
		s.WriteString(initInfoStyle.Render("What would you like to do?"))
		s.WriteString("\n\n")

		options := []struct {
			label string
			desc  string
		}{
			{"Ignore", "Keep existing configuration"},
			{"Override", "Replace with new configuration"},
		}

		for i, opt := range options {
			cursor := "  "
			if m.cursor == i {
				cursor = initSelectedStyle.Render("→ ")
				s.WriteString(cursor)
				s.WriteString(initSelectedStyle.Render(opt.label))
				s.WriteString(initNormalStyle.Render(" - " + opt.desc))
			} else {
				s.WriteString(cursor)
				s.WriteString(initNormalStyle.Render(opt.label))
				s.WriteString(initNormalStyle.Render(" - " + opt.desc))
			}
			s.WriteString("\n")
		}

		s.WriteString("\n")
		s.WriteString(initNormalStyle.Render("↑/↓: navigate • enter: select • q/esc: cancel"))
	} else if m.step == 1 {
		// Step 2: Confirm override
		s.WriteString("\n")
		s.WriteString(initDangerStyle.Render("⚠  WARNING: This will delete your existing configuration!"))
		s.WriteString("\n\n")
		s.WriteString(initWarningStyle.Render("All stored passwords and settings will be permanently lost."))
		s.WriteString("\n")
		s.WriteString(initWarningStyle.Render("This action cannot be undone."))
		s.WriteString("\n\n")
		s.WriteString(initInfoStyle.Render("Are you sure you want to continue?"))
		s.WriteString("\n\n")

		options := []struct {
			label string
			style lipgloss.Style
		}{
			{"Yes, override", initDangerStyle},
			{"No, cancel", initNormalStyle},
		}

		for i, opt := range options {
			cursor := "  "
			if m.cursor == i {
				cursor = initSelectedStyle.Render("→ ")
				s.WriteString(cursor)
				s.WriteString(opt.style.Render(opt.label))
			} else {
				s.WriteString(cursor)
				s.WriteString(opt.style.Render(opt.label))
			}
			s.WriteString("\n")
		}

		s.WriteString("\n")
		s.WriteString(initNormalStyle.Render("↑/↓: navigate • enter: confirm • q/esc: cancel"))
	}

	s.WriteString("\n")
	return s.String()
}

// RunInitTUI shows the init prompt and returns the user's choice
func RunInitTUI() (InitChoice, error) {
	model := NewInitTUI()
	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		// If TUI fails (no TTY), return cancel so caller can use fallback
		return ChoiceCancel, err
	}

	finalModel := result.(initModel)
	return InitChoice(finalModel.selected), nil
}
