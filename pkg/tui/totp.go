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

	totpQRStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#000000")).
			Padding(1)
)

func NewTOTPSetupTUI(accountName string) *totpModel {
	return &totpModel{
		step:        0,
		accountName: accountName,
		width:       80,
		height:      24,
	}
}

func (m totpModel) Init() tea.Cmd {
	return nil
}

func (m totpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.success {
			return m, tea.Quit
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.err != nil || m.step == 0 {
				return m, tea.Quit
			}

		case "esc":
			if m.step > 0 {
				m.step = 0
				m.err = nil
				m.copiedMessage = ""
				return m, nil
			}
			return m, tea.Quit

		case "c":
			if m.step == 1 && m.serverURL != "" {
				if err := copyToClipboardTOTP(m.serverURL); err == nil {
					m.copiedMessage = "✓ Copied URL to clipboard"
				} else {
					m.copiedMessage = fmt.Sprintf("✗ Failed to copy: %v", err)
				}
			}

		case "enter":
			if m.step == 0 {
				if m.accountName == "" {
					m.accountName = "user"
				}

				key, err := mfa.GenerateTOTPSecret(m.accountName)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.key = key

				m.qrServer = server.NewQRServer()
				url, err := m.qrServer.Start(m.key.Secret, m.key.AccountName, m.key.Issuer)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.serverURL = url

				if err := server.OpenBrowser(url); err != nil {
					m.err = fmt.Errorf("failed to open browser: %w (please visit %s manually)", err, url)
				}

				m.step = 1
			} else if m.step == 1 {
				m.step = 2
			} else if m.step == 2 {
				if len(m.codeInput) != 6 {
					m.err = fmt.Errorf("code must be 6 digits")
					return m, nil
				}

				if !mfa.ValidateTOTP(m.key.Secret, m.codeInput) {
					m.err = fmt.Errorf("invalid TOTP code, please try again")
					m.codeInput = ""
					return m, nil
				}

				if err := config.SaveTOTPSecret(m.key.Secret); err != nil {
					m.err = err
					return m, nil
				}

				m.success = true
				return m, tea.Quit
			}

		case "backspace":
			if m.step == 0 && len(m.accountName) > 0 {
				m.accountName = m.accountName[:len(m.accountName)-1]
			} else if m.step == 2 && len(m.codeInput) > 0 {
				m.codeInput = m.codeInput[:len(m.codeInput)-1]
				m.err = nil
			}

		default:
			if len(msg.String()) == 1 {
				if m.step == 0 {
					m.accountName += msg.String()
				} else if m.step == 2 && len(m.codeInput) < 6 {
					if msg.String()[0] >= '0' && msg.String()[0] <= '9' {
						m.codeInput += msg.String()
						m.err = nil
					}
				}
			}
		}
	}

	return m, nil
}

func (m totpModel) View() string {
	if m.success {
		return totpSuccessStyle.Render("✓ TOTP authentication enabled successfully!\n\nPress any key to exit...")
	}

	if m.err != nil && m.step == 0 {
		return totpErrorStyle.Render(fmt.Sprintf("✗ Error: %v\n\nPress 'q' to exit", m.err))
	}

	var s strings.Builder

	switch m.step {
	case 0:
		s.WriteString(totpTitleStyle.Render("TOTP Setup - Step 1"))
		s.WriteString("\n\n")

		s.WriteString(totpLabelStyle.Render("Account Name: "))
		if m.accountName == "" {
			s.WriteString(totpInputStyle.Render("user▋"))
		} else {
			s.WriteString(totpInputStyle.Render(m.accountName + "▋"))
		}
		s.WriteString("\n\n")

		s.WriteString(totpNormalStyle.Render("This will be shown in your authenticator app"))
		s.WriteString("\n\n")
		s.WriteString(totpNormalStyle.Render("Press Enter to continue • q: quit"))

	case 1:
		s.WriteString(totpTitleStyle.Render("TOTP Setup - Step 2: Scan QR Code"))
		s.WriteString("\n\n")

		if m.err != nil {
			s.WriteString(totpErrorStyle.Render(fmt.Sprintf("✗ %v", m.err)))
			s.WriteString("\n\n")
		} else {
			s.WriteString(totpSuccessStyle.Render("✓ Browser opened with QR code"))
			s.WriteString("\n\n")
		}

		s.WriteString(totpNormalStyle.Render("QR code available at: "))
		s.WriteString(totpValueStyle.Render(m.serverURL))
		s.WriteString("\n\n")

		s.WriteString(totpLabelStyle.Render("Manual entry:"))
		s.WriteString("\n")
		s.WriteString(totpNormalStyle.Render("Secret: "))
		s.WriteString(totpValueStyle.Render(m.key.Secret))
		s.WriteString("\n")
		s.WriteString(totpNormalStyle.Render("Issuer: "))
		s.WriteString(totpValueStyle.Render(m.key.Issuer))
		s.WriteString("\n")
		s.WriteString(totpNormalStyle.Render("Account: "))
		s.WriteString(totpValueStyle.Render(m.key.AccountName))
		s.WriteString("\n\n")

		if m.copiedMessage != "" {
			if strings.HasPrefix(m.copiedMessage, "✓") {
				s.WriteString(totpSuccessStyle.Render(m.copiedMessage))
			} else {
				s.WriteString(totpErrorStyle.Render(m.copiedMessage))
			}
			s.WriteString("\n\n")
		}

		s.WriteString(totpNormalStyle.Render("c: copy URL • enter: continue • esc: back • q: quit"))

	case 2:
		s.WriteString(totpTitleStyle.Render("TOTP Setup - Step 3: Verify"))
		s.WriteString("\n\n")

		s.WriteString(totpNormalStyle.Render("Enter the 6-digit code from your authenticator app:"))
		s.WriteString("\n\n")

		s.WriteString(totpLabelStyle.Render("Code: "))
		if len(m.codeInput) == 0 {
			s.WriteString(totpInputStyle.Render("______▋"))
		} else {
			display := m.codeInput + strings.Repeat("_", 6-len(m.codeInput))
			s.WriteString(totpInputStyle.Render(display + "▋"))
		}
		s.WriteString("\n")

		if m.err != nil {
			s.WriteString("\n")
			s.WriteString(totpErrorStyle.Render(fmt.Sprintf("✗ %v", m.err)))
		}

		s.WriteString("\n\n")
		s.WriteString(totpNormalStyle.Render("Enter 6 digits • backspace: delete • esc: back • q: quit"))
	}

	return s.String()
}

func copyToClipboardTOTP(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip, xsel, or wl-clipboard)")
		}
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := stdin.Write([]byte(text)); err != nil {
		return err
	}

	if err := stdin.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

func RunTOTPSetupTUI(accountName string) error {
	p := tea.NewProgram(
		NewTOTPSetupTUI(accountName),
	)
	model, err := p.Run()
	if err != nil {
		return err
	}

	m := model.(totpModel)

	if m.qrServer != nil {
		m.qrServer.Stop()
	}

	if m.success {
		return nil
	}

	return fmt.Errorf("setup cancelled")
}
