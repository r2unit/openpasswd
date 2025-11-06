package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/config"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/models"
)

type modernModel struct {
	db                *database.DB
	encryptor         *crypto.Encryptor
	passwords         []*models.Password
	filteredPasswords []*models.Password
	recentPasswords   []*models.Password
	commonPasswords   []string
	searchInput       string
	cursor            int
	width             int
	height            int
	err               error
	showDetails       bool
	selectedPass      *models.Password
	colors            *config.ColorScheme
}

var (
	searchBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5FAFFF")).
			Padding(0, 3).
			MarginBottom(1)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#585858")).
			Padding(1, 3).
			Width(30)

	mainBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#585858")).
			Padding(1, 3)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#5FAFFF")).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8")).
			Padding(0, 1)

	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFAF00")).
				MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585858")).
			MarginTop(1)
)

func NewModernTUI(db *database.DB, salt []byte, passphrase string) *modernModel {
	encryptor := crypto.NewEncryptor(passphrase, salt)

	passwords, _ := db.ListPasswords()

	common := []string{
		"GitHub", "Gmail", "Google", "Twitter", "Facebook",
		"LinkedIn", "Instagram", "Netflix", "Amazon", "Spotify",
	}

	recent := getRecentPasswords(passwords, 10)

	colors, _ := config.LoadColorScheme()

	return &modernModel{
		db:                db,
		encryptor:         encryptor,
		passwords:         passwords,
		filteredPasswords: passwords,
		recentPasswords:   recent,
		commonPasswords:   common,
		searchInput:       "",
		cursor:            0,
		width:             80,
		height:            24,
		colors:            colors,
	}
}

func getRecentPasswords(passwords []*models.Password, limit int) []*models.Password {
	if len(passwords) == 0 {
		return []*models.Password{}
	}

	sorted := make([]*models.Password, len(passwords))
	copy(sorted, passwords)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
	})

	if len(sorted) > limit {
		return sorted[:limit]
	}
	return sorted
}

func (m modernModel) Init() tea.Cmd {
	return nil
}

func (m modernModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.showDetails {
				return m, tea.Quit
			}
			m.showDetails = false
			return m, nil

		case "esc":
			if m.showDetails {
				m.showDetails = false
			} else if m.searchInput != "" {
				m.searchInput = ""
				m.filteredPasswords = m.passwords
				m.cursor = 0
			} else {
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filteredPasswords)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.filteredPasswords) > 0 && m.cursor < len(m.filteredPasswords) {
				m.selectedPass = m.filteredPasswords[m.cursor]
				m.showDetails = true
			}

		case "ctrl+n", "n":
			return m, tea.Quit

		case "backspace":
			if len(m.searchInput) > 0 {
				m.searchInput = m.searchInput[:len(m.searchInput)-1]
				m.filterPasswords()
				m.cursor = 0
			}

		default:
			if len(msg.String()) == 1 && !m.showDetails {
				m.searchInput += msg.String()
				m.filterPasswords()
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m *modernModel) filterPasswords() {
	if m.searchInput == "" {
		m.filteredPasswords = m.passwords
		return
	}

	query := strings.ToLower(m.searchInput)
	filtered := []*models.Password{}

	for _, p := range m.passwords {
		decryptedName := p.Name
		if name, err := m.encryptor.Decrypt(p.Name); err == nil {
			decryptedName = name
		}

		decryptedUsername := p.Username
		if p.Username != "" {
			if username, err := m.encryptor.Decrypt(p.Username); err == nil {
				decryptedUsername = username
			}
		}

		decryptedURL := p.URL
		if p.URL != "" {
			if url, err := m.encryptor.Decrypt(p.URL); err == nil {
				decryptedURL = url
			}
		}

		if strings.Contains(strings.ToLower(decryptedName), query) ||
			strings.Contains(strings.ToLower(decryptedUsername), query) ||
			strings.Contains(strings.ToLower(decryptedURL), query) {
			filtered = append(filtered, p)
		}
	}

	m.filteredPasswords = filtered
}

func (m modernModel) View() string {
	if m.showDetails {
		return m.renderDetails()
	}

	var content strings.Builder

	content.WriteString("\n")

	mainWidth := m.width - 35
	if mainWidth < 40 {
		mainWidth = 40
	}

	sidebar := m.renderSidebar()
	searchBar := m.renderSearchBar(mainWidth)
	mainBox := m.renderMainBox(mainWidth)

	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		" ",
		searchBar,
	)

	content.WriteString(topRow)
	content.WriteString("\n")

	mainRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Repeat(" ", 32),
		mainBox,
	)

	content.WriteString(mainRow)

	contentHeight := lipgloss.Height(content.String())
	spacerHeight := m.height - contentHeight - 4
	if spacerHeight < 0 {
		spacerHeight = 0
	}

	for i := 0; i < spacerHeight; i++ {
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(m.renderFooter())

	return content.String()
}

func (m modernModel) renderSearchBar(width int) string {
	searchText := m.searchInput
	if searchText == "" {
		searchText = "Search passwords..."
	}

	barStyle := searchBarStyle.Copy().Width(width - 4)
	return barStyle.Render(fmt.Sprintf("🔍 %s", searchText))
}

func (m modernModel) renderSidebar() string {
	var s strings.Builder

	s.WriteString(sidebarTitleStyle.Render("Recent"))
	s.WriteString("\n\n")

	if len(m.recentPasswords) == 0 {
		s.WriteString(normalItemStyle.Render("No recent passwords"))
	} else {
		displayCount := len(m.recentPasswords)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			if i > 0 {
				s.WriteString("\n")
			}
			p := m.recentPasswords[i]
			timeAgo := formatTimeAgo(p.UpdatedAt)

			decryptedName := p.Name
			if name, err := m.encryptor.Decrypt(p.Name); err == nil {
				decryptedName = name
			}

			s.WriteString(normalItemStyle.Render(fmt.Sprintf("• %s", truncate(decryptedName, 24))))
			s.WriteString("\n")
			s.WriteString(normalItemStyle.Render(fmt.Sprintf("  %s", timeAgo)))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(sidebarTitleStyle.Render("Common"))
	s.WriteString("\n\n")

	for i, name := range m.commonPasswords {
		if i > 0 {
			s.WriteString("\n")
		}

		exists := false
		for _, p := range m.passwords {
			decryptedName := p.Name
			if name, err := m.encryptor.Decrypt(p.Name); err == nil {
				decryptedName = name
			}
			if strings.Contains(strings.ToLower(decryptedName), strings.ToLower(name)) {
				exists = true
				break
			}
		}

		if exists {
			s.WriteString(normalItemStyle.Render(fmt.Sprintf("✓ %s", name)))
		} else {
			s.WriteString(normalItemStyle.Render(fmt.Sprintf("  %s", name)))
		}
	}

	contentHeight := lipgloss.Height(s.String())
	availableHeight := m.height - 10
	if availableHeight < contentHeight {
		availableHeight = contentHeight
	}

	sidebarWithHeight := sidebarStyle.Copy().Height(availableHeight)
	return sidebarWithHeight.Render(s.String())
}

func (m modernModel) renderMainBox(width int) string {
	var s strings.Builder

	s.WriteString(headerStyle.Render(fmt.Sprintf("Passwords (%d)", len(m.filteredPasswords))))
	s.WriteString("\n\n")

	if len(m.filteredPasswords) == 0 {
		s.WriteString(normalItemStyle.Render("No passwords found"))
	} else {
		maxVisible := 12
		start := m.cursor
		if start > len(m.filteredPasswords)-maxVisible {
			start = len(m.filteredPasswords) - maxVisible
		}
		if start < 0 {
			start = 0
		}

		end := start + maxVisible
		if end > len(m.filteredPasswords) {
			end = len(m.filteredPasswords)
		}

		for i := start; i < end; i++ {
			if i > start {
				s.WriteString("\n")
			}
			p := m.filteredPasswords[i]

			decryptedName := p.Name
			if name, err := m.encryptor.Decrypt(p.Name); err == nil {
				decryptedName = name
			}

			decryptedUsername := p.Username
			if p.Username != "" {
				if username, err := m.encryptor.Decrypt(p.Username); err == nil {
					decryptedUsername = username
				}
			}

			line := fmt.Sprintf("%-20s  %-20s", truncate(decryptedName, 20), truncate(decryptedUsername, 20))

			if i == m.cursor {
				s.WriteString(selectedItemStyle.Render("→ " + line))
			} else {
				s.WriteString(normalItemStyle.Render("  " + line))
			}
		}
	}

	boxStyle := mainBoxStyle.Copy().Width(width - 4)
	return boxStyle.Render(s.String())
}

func (m modernModel) renderDetails() string {
	if m.selectedPass == nil {
		return "No password selected"
	}

	var s strings.Builder

	s.WriteString(headerStyle.Render("Password Details"))
	s.WriteString("\n\n")

	decrypted, err := m.encryptor.Decrypt(m.selectedPass.Password)
	if err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error decrypting: %v", err)))
	} else {
		details := []struct {
			label string
			value string
		}{
			{"ID", fmt.Sprintf("%d", m.selectedPass.ID)},
			{"Name", m.selectedPass.Name},
			{"Username", m.selectedPass.Username},
			{"Password", decrypted},
			{"URL", m.selectedPass.URL},
			{"Notes", m.selectedPass.Notes},
			{"Created", m.selectedPass.CreatedAt.Format("2006-01-02 15:04:05")},
			{"Updated", m.selectedPass.UpdatedAt.Format("2006-01-02 15:04:05")},
		}

		for _, d := range details {
			s.WriteString(headerStyle.Render(d.label + ":"))
			s.WriteString(" ")
			s.WriteString(normalItemStyle.Render(d.value))
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press 'q' or 'esc' to go back"))

	return mainBoxStyle.Render(s.String())
}

func (m modernModel) renderFooter() string {
	help := "↑/↓: navigate • enter: view details • n: new password • q: quit • esc: back/clear"
	version := "openpass v1.0.0"

	if m.width < 80 {
		help = "↑/↓: nav • enter: view • n: new • q: quit"
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#585858"))

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3A3A3A")).
		Bold(true).
		Width(30).
		Align(lipgloss.Center)

	versionText := versionStyle.Render(version)

	centerWidth := m.width - 35
	if centerWidth < 40 {
		centerWidth = 40
	}
	if centerWidth > len(help)+10 {
		centerWidth = len(help) + 10
	}

	centerHelp := lipgloss.NewStyle().
		Width(centerWidth).
		Align(lipgloss.Center).
		Render(helpStyle.Render(help))

	spacerWidth := m.width - 30 - centerWidth
	if spacerWidth < 1 {
		spacerWidth = 1
	}

	spacer := strings.Repeat(" ", spacerWidth)

	footerLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		versionText,
		spacer,
		centerHelp,
	)

	return footerLine
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("Jan 2")
	}
}

func RunModernTUI(db *database.DB, salt []byte, passphrase string) error {
	p := tea.NewProgram(
		NewModernTUI(db, salt, passphrase),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
