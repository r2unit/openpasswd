package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VersionInfo contains version comparison data
type VersionInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	GitCommit       string
	BuildDate       string
	GoVersion       string
	Platform        string
	ReleaseNotes    string
	ReleaseURL      string
}

type versionModel struct {
	info     VersionInfo
	width    int
	height   int
	cursor   int // 0 = Upgrade, 1 = Skip
	quitting bool
	choice   bool // true = upgrade, false = skip
}

func NewVersionModel(info VersionInfo) *versionModel {
	return &versionModel{
		info:   info,
		width:  80,
		height: 24,
		cursor: 0,
	}
}

func (m versionModel) Init() tea.Cmd {
	return nil
}

func (m versionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			m.choice = false
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}

		case "left", "h":
			m.cursor = 0

		case "right", "l":
			m.cursor = 1

		case "enter", " ":
			m.quitting = true
			m.choice = m.cursor == 0 // true if "Upgrade" selected
			return m, tea.Quit

		case "y":
			m.quitting = true
			m.choice = true
			return m, tea.Quit

		case "n":
			m.quitting = true
			m.choice = false
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m versionModel) View() string {
	var s strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(0, 1).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF00")).
		Padding(1, 2).
		Width(70)

	versionBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#5F87FF")).
		Padding(0, 1).
		Width(30).
		Align(lipgloss.Center)

	currentVersionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	newVersionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FAFFF"))

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#5F87FF")).
		Padding(0, 3).
		Bold(true)

	selectedButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color("#00FF00")).
		Padding(0, 3).
		Bold(true)

	// Header
	s.WriteString("\n")
	if m.info.UpdateAvailable {
		s.WriteString(titleStyle.Render("⚡ VERSION CHECK - UPDATE AVAILABLE ⚡"))
	} else {
		s.WriteString(titleStyle.Render("✓ VERSION CHECK - UP TO DATE ✓"))
	}
	s.WriteString("\n\n")

	// Version comparison boxes
	if m.info.UpdateAvailable {
		// ASCII art for update available
		art := `
    ╔═══════════════╗         ╔═══════════════╗
    ║   CURRENT     ║  ═══►   ║     LATEST    ║
    ╚═══════════════╝         ╚═══════════════╝
`
		artStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Align(lipgloss.Center)
		s.WriteString(artStyle.Render(art))
		s.WriteString("\n")

		// Version boxes side by side
		currentBox := versionBoxStyle.Render(
			currentVersionStyle.Render(fmt.Sprintf("v%s", m.info.CurrentVersion)) + "\n" +
				dimStyle.Render("(installed)"),
		)

		latestBox := versionBoxStyle.Copy().
			BorderForeground(lipgloss.Color("#00FF00")).
			Render(
				newVersionStyle.Render(fmt.Sprintf("v%s", m.info.LatestVersion)) + "\n" +
					dimStyle.Render("(available)"),
			)

		versions := lipgloss.JoinHorizontal(lipgloss.Top,
			currentBox,
			"  →  ",
			latestBox,
		)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(versions))
		s.WriteString("\n\n")
	} else {
		// Up to date
		art := `
    ╔═══════════════════════════════════╗
    ║   ✓ RUNNING LATEST VERSION ✓     ║
    ╚═══════════════════════════════════╝
`
		artStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Align(lipgloss.Center)
		s.WriteString(artStyle.Render(art))
		s.WriteString("\n")

		versionBox := versionBoxStyle.Copy().
			BorderForeground(lipgloss.Color("#00FF00")).
			Width(40).
			Render(
				newVersionStyle.Render(fmt.Sprintf("v%s", m.info.CurrentVersion)) + "\n" +
					dimStyle.Render("(up to date)"),
			)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(versionBox))
		s.WriteString("\n\n")
	}

	// Build information box
	var content strings.Builder
	content.WriteString(labelStyle.Render("Build Information:") + "\n\n")
	content.WriteString(fmt.Sprintf("  Git Commit:  %s\n", dimStyle.Render(m.info.GitCommit)))
	content.WriteString(fmt.Sprintf("  Build Date:  %s\n", dimStyle.Render(m.info.BuildDate)))
	content.WriteString(fmt.Sprintf("  Go Version:  %s\n", dimStyle.Render(m.info.GoVersion)))
	content.WriteString(fmt.Sprintf("  Platform:    %s\n", dimStyle.Render(m.info.Platform)))

	if m.info.UpdateAvailable && m.info.ReleaseNotes != "" {
		content.WriteString("\n")
		content.WriteString(labelStyle.Render("What's New:") + "\n\n")
		// Truncate release notes if too long
		notes := m.info.ReleaseNotes
		if len(notes) > 200 {
			notes = notes[:200] + "..."
		}
		content.WriteString(dimStyle.Render("  " + strings.ReplaceAll(notes, "\n", "\n  ")))
		content.WriteString("\n")
	}

	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(boxStyle.Render(content.String())))
	s.WriteString("\n\n")

	// Action buttons (only if update available)
	if m.info.UpdateAvailable {
		upgradeBtn := "[ Upgrade Now ]"
		skipBtn := "[ Skip ]"

		if m.cursor == 0 {
			upgradeBtn = selectedButtonStyle.Render(upgradeBtn)
			skipBtn = buttonStyle.Render(skipBtn)
		} else {
			upgradeBtn = buttonStyle.Render(upgradeBtn)
			skipBtn = selectedButtonStyle.Render(skipBtn)
		}

		buttons := lipgloss.JoinHorizontal(lipgloss.Top, upgradeBtn, "  ", skipBtn)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(buttons))
		s.WriteString("\n\n")

		// Navigation hints
		hints := dimStyle.Render("← / h: Upgrade  |  → / l: Skip  |  Enter: Confirm  |  q: Cancel")
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(hints))
	} else {
		footer := dimStyle.Render("Press 'q' or ESC to exit")
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(footer))
	}

	s.WriteString("\n")

	return s.String()
}

// RunVersionTUI shows the version comparison TUI
// Returns true if user chose to upgrade, false otherwise
func RunVersionTUI(info VersionInfo) (bool, error) {
	model := NewVersionModel(info)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	m, ok := finalModel.(versionModel)
	if !ok {
		return false, nil
	}

	return m.choice, nil
}

// RunVersionCheckBanner shows a non-intrusive version check banner
// This is for startup checks, not interactive
func RunVersionCheckBanner(currentVersion, latestVersion string) {
	if latestVersion == "" || currentVersion == latestVersion {
		return
	}

	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFD700")).
		Foreground(lipgloss.Color("#FFD700")).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(1)

	banner := fmt.Sprintf(
		"⚠ Update available: v%s → v%s | Run 'openpasswd upgrade' to update",
		currentVersion,
		latestVersion,
	)

	fmt.Println(bannerStyle.Render(banner))
}
