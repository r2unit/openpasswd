package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/config"
	"github.com/r2unit/openpasswd/pkg/mfa"
	"github.com/r2unit/openpasswd/pkg/server"
)

type totpModel struct {
	step          int
	accountName   string
	key           *mfa.TOTPKey
	serverURL     string
	codeInput     string
	err           error
	success       bool
	width         int
	height        int
	qrServer      *server.QRServer
	copiedMessage string
}

var (
	totpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFAF00")).
			MarginBottom(1)

	totpLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAF00")).
			Bold(true)

	totpValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	totpNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8"))

	totpInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	totpErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			Bold(true)

	totpSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D75F")).
				Bold(true)

