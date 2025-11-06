package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/models"
)

type view int

const (
	menuView view = iota
	listView
	viewDetailsView
	addView
	searchView
	updateView
	deleteView
)

type model struct {
	db           *database.DB
	encryptor    *crypto.Encryptor
	currentView  view
	cursor       int
	passwords    []*models.Password
	selectedPass *models.Password
	input        string
	err          error
	message      string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1)

	menuStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#585858")).
			Padding(1, 2).
			MarginTop(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#585858")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#D70000", Dark: "#FF5F5F"}).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#00AF00", Dark: "#00D75F"}).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#D78700", Dark: "#FFAF00"}).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#0087D7", Dark: "#5FAFFF"}).
			Bold(true)
)

func NewBubbleTea(db *database.DB, salt []byte, passphrase string) *model {
	encryptor := crypto.NewEncryptor(passphrase, salt)
	return &model{
		db:          db,
		encryptor:   encryptor,
		currentView: menuView,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			m.currentView = menuView
			m.message = ""
			m.err = nil
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxCursor := 0
			switch m.currentView {
			case menuView:
				maxCursor = 5
			case listView:
				maxCursor = len(m.passwords) - 1
			}
			if m.cursor < maxCursor {
				m.cursor++
			}

		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			if m.currentView == menuView {
				num := int(msg.String()[0] - '0')
				if num >= 1 && num <= 6 {
					m.cursor = num - 1
					return m.handleEnter()
				}
			} else if m.currentView == listView {
				num := int(msg.String()[0] - '0')
				if num >= 1 && num <= len(m.passwords) && num <= 9 {
					m.cursor = num - 1
					return m.handleEnter()
				}
			}

		case "enter":
			return m.handleEnter()
		}
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case menuView:
		switch m.cursor {
		case 0:
			passwords, err := m.db.ListPasswords()
			if err != nil {
				m.err = err
				return m, nil
			}
			m.passwords = passwords
			m.currentView = listView
			m.cursor = 0
		case 1:
			m.currentView = addView
			m.message = "Add feature requires interactive input. Use 'openpass add' command."
		case 2:
			m.currentView = searchView
			m.message = "Search feature requires interactive input. Use TUI menu option 4."
		case 3:
			m.currentView = updateView
			m.message = "Update feature requires interactive input. Use TUI menu option 5."
		case 4:
			m.currentView = deleteView
			m.message = "Delete feature requires interactive input. Use TUI menu option 6."
		case 5:
			return m, tea.Quit
		}
	case listView:
		if m.cursor < len(m.passwords) {
			m.selectedPass = m.passwords[m.cursor]
			m.currentView = viewDetailsView
		}
	case viewDetailsView:
		m.currentView = listView
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("OpenPasswd - Password Manager"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\n")
	}

	if m.message != "" {
		s.WriteString(successStyle.Render(m.message))
		s.WriteString("\n\n")
	}

	switch m.currentView {
	case menuView:
		s.WriteString(m.renderMenu())
	case listView:
		s.WriteString(m.renderList())
	case viewDetailsView:
		s.WriteString(m.renderDetails())
	default:
		s.WriteString(m.renderMessage())
	}

	s.WriteString("\n\n")
	s.WriteString(normalStyle.Render("Press 'q' to quit, 'esc' to go back, ↑/↓ or numbers to navigate, enter to select"))
	s.WriteString("\n")

	return s.String()
}

func (m model) renderMenu() string {
	menu := []string{
		"List all passwords",
		"Add new password",
		"Search passwords",
		"Update password",
		"Delete password",
		"Exit",
	}

	var s strings.Builder

	for i, item := range menu {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, item)))
		} else {
			s.WriteString(normalStyle.Render(fmt.Sprintf("%s %d. %s", cursor, i+1, item)))
		}
		s.WriteString("\n")
	}

	return s.String()
}

func (m model) renderList() string {
	var s strings.Builder

	if len(m.passwords) == 0 {
		s.WriteString(normalStyle.Render("No passwords stored yet."))
		return s.String()
	}

	s.WriteString(titleStyle.Render("Stored Passwords"))
	s.WriteString("\n\n")

	for i, p := range m.passwords {
		cursor := " "
		line := fmt.Sprintf("%s [%d] %s (%s)", cursor, p.ID, p.Name, p.Username)

		if m.cursor == i {
			cursor = ">"
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(normalStyle.Render(line))
		}
		s.WriteString("\n")
	}

	return s.String()
}

func (m model) renderDetails() string {
	if m.selectedPass == nil {
		return normalStyle.Render("No password selected")
	}

	var s strings.Builder

	decrypted, err := m.encryptor.Decrypt(m.selectedPass.Password)
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error decrypting password: %v", err))
	}

	s.WriteString(titleStyle.Render("Password Details"))
	s.WriteString("\n\n")

	details := []string{
		fmt.Sprintf("ID:       %d", m.selectedPass.ID),
		fmt.Sprintf("Name:     %s", m.selectedPass.Name),
		fmt.Sprintf("Username: %s", m.selectedPass.Username),
		fmt.Sprintf("Password: %s", decrypted),
		fmt.Sprintf("URL:      %s", m.selectedPass.URL),
		fmt.Sprintf("Notes:    %s", m.selectedPass.Notes),
		fmt.Sprintf("Created:  %s", m.selectedPass.CreatedAt.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Updated:  %s", m.selectedPass.UpdatedAt.Format("2006-01-02 15:04:05")),
	}

	for _, detail := range details {
		s.WriteString(normalStyle.Render(detail))
		s.WriteString("\n")
	}

	return s.String()
}

func (m model) renderMessage() string {
	var s strings.Builder
	s.WriteString(normalStyle.Render("This feature requires interactive input."))
	s.WriteString("\n")
	s.WriteString(normalStyle.Render("Please use the command-line interface or the original TUI."))
	return s.String()
}

func RunBubbleTea(db *database.DB, salt []byte, passphrase string) error {
	p := tea.NewProgram(NewBubbleTea(db, salt, passphrase))
	_, err := p.Run()
	return err
}
