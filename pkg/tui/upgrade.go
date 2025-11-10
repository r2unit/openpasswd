package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Upgrade stages
type upgradeStage int

const (
	stageInitializing upgradeStage = iota
	stageConnecting
	stageBeaming
	stageVerifying
	stageInstalling
	stageComplete
	stageError
)

type upgradeModel struct {
	stage       upgradeStage
	progress    int
	maxProgress int
	message     string
	details     string
	version     string
	width       int
	height      int
	spinner     int
	err         error
}

type upgradeTickMsg time.Time
type upgradeCompleteMsg struct{}
type upgradeErrorMsg struct{ err error }

var upgradeSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var beamFrames = []string{
	"[          ]",
	"[>         ]",
	"[=>        ]",
	"[==>       ]",
	"[===>      ]",
	"[====>     ]",
	"[=====>    ]",
	"[======>   ]",
	"[=======>  ]",
	"[========> ]",
	"[=========>]",
}

func NewUpgradeModel(version string) *upgradeModel {
	return &upgradeModel{
		stage:       stageInitializing,
		progress:    0,
		maxProgress: 100,
		version:     version,
		width:       80,
		height:      24,
		spinner:     0,
	}
}

func (m upgradeModel) Init() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return upgradeTickMsg(t)
	})
}

func (m upgradeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case upgradeTickMsg:
		m.spinner = (m.spinner + 1) % len(upgradeSpinnerFrames)

		// Simulate upgrade progress
		m.progress += 2

		// Update stage based on progress
		switch {
		case m.progress < 15:
			m.stage = stageInitializing
			m.message = "Initializing upgrade sequence..."
			m.details = "Preparing quantum entanglement matrix"
		case m.progress < 30:
			m.stage = stageConnecting
			m.message = "Establishing connection to GitHub..."
			m.details = fmt.Sprintf("DNS lookup: api.github.com → %s", "140.82.121.6")
		case m.progress < 70:
			m.stage = stageBeaming
			m.message = "Beaming data from GitHub servers..."
			beamIdx := (m.progress / 4) % len(beamFrames)
			m.details = fmt.Sprintf("Transporter buffer: %s %.1f%%", beamFrames[beamIdx], float64(m.progress)*1.4)
		case m.progress < 85:
			m.stage = stageVerifying
			m.message = "Verifying cryptographic signatures..."
			m.details = "SHA256: c0ffee...deadbeef (VALID ✓)"
		case m.progress < 95:
			m.stage = stageInstalling
			m.message = "Installing new binary..."
			m.details = "chmod +x openpasswd && mv -f"
		case m.progress >= 100:
			m.stage = stageComplete
			m.message = "Upgrade complete!"
			m.details = fmt.Sprintf("Successfully upgraded to v%s", m.version)
			return m, tea.Quit
		}

		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return upgradeTickMsg(t)
		})

	case upgradeCompleteMsg:
		m.stage = stageComplete
		return m, tea.Quit

	case upgradeErrorMsg:
		m.stage = stageError
		m.err = msg.err
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m upgradeModel) View() string {
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
		Width(60)

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FAFFF")).
		Italic(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F")).
		Bold(true)

	// Header
	s.WriteString("\n")
	s.WriteString(titleStyle.Render("⚡ OPENPASSWD UPGRADE SYSTEM v2.0 ⚡"))
	s.WriteString("\n\n")

	// ASCII art
	art := `
    ╔═══════════════════════════╗
    ║  [GITHUB] ═══► [CLIENT]  ║
    ╚═══════════════════════════╝
`
	artStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Align(lipgloss.Center)
	s.WriteString(artStyle.Render(art))
	s.WriteString("\n")

	// Main content box
	var content strings.Builder

	// Status
	spinner := upgradeSpinnerFrames[m.spinner]
	statusColor := "#00FF00"
	if m.stage == stageError {
		statusColor = "#FF5F5F"
		spinner = "✗"
	} else if m.stage == stageComplete {
		spinner = "✓"
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true)

	content.WriteString(statusStyle.Render(fmt.Sprintf("%s  %s", spinner, m.message)))
	content.WriteString("\n\n")

	// Progress bar
	if m.stage != stageComplete && m.stage != stageError {
		progressBar := renderProgressBar(m.progress, m.maxProgress, 50)
		content.WriteString(progressStyle.Render(progressBar))
		content.WriteString("\n")
		content.WriteString(progressStyle.Render(fmt.Sprintf("Progress: %d%%", m.progress)))
		content.WriteString("\n\n")
	}

	// Details
	if m.details != "" {
		if m.stage == stageError {
			content.WriteString(errorStyle.Render("ERROR: " + m.details))
		} else {
			content.WriteString(detailStyle.Render("» " + m.details))
		}
		content.WriteString("\n")
	}

	// System info (nerdy details)
	if m.stage == stageBeaming {
		content.WriteString("\n")
		content.WriteString(dimStyle.Render("Protocol: HTTPS/2 | TLS 1.3"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(fmt.Sprintf("Packets: %d | Latency: %.1fms", m.progress*10, float64(m.progress)/5+10)))
	}

	s.WriteString(boxStyle.Render(content.String()))
	s.WriteString("\n\n")

	// Footer
	if m.stage == stageComplete {
		footer := successStyle.Render("🎉 Upgrade successful! Restart to use new version.")
		s.WriteString(footer)
	} else if m.stage == stageError {
		footer := errorStyle.Render("Upgrade failed. Press any key to exit.")
		s.WriteString(footer)
	} else {
		footer := dimStyle.Render("Press 'q' to cancel")
		s.WriteString(footer)
	}

	s.WriteString("\n")

	return s.String()
}

func renderProgressBar(current, max, width int) string {
	if max == 0 {
		return ""
	}

	percentage := float64(current) / float64(max)
	filled := int(percentage * float64(width))

	if filled > width {
		filled = width
	}

	bar := strings.Builder{}
	bar.WriteString("│")

	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	bar.WriteString("│")
	return bar.String()
}

// RunUpgradeTUI runs the upgrade TUI
func RunUpgradeTUI(version string) error {
	model := NewUpgradeModel(version)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}

// Simulation helpers for testing
func SimulateUpgradeProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return upgradeTickMsg(t)
	})
}
