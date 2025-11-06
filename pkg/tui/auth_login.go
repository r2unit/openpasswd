package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/auth"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
)

type authLoginModel struct {
	db               *database.DB
	salt             []byte
	masterPass       string
	providers        []auth.Provider
	cursor           int
	step             int // 0: select provider, 1: enter credentials, 2: syncing, 3: done
	selectedProvider auth.Provider
	credentialInputs map[string]string
	currentField     int
	errorMsg         string
	successMsg       string
	syncedCount      int
	width            int
	height           int
}

var (
	authTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFAF00")).
			MarginBottom(1)

	authSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	authNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8"))

	authLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAF00")).
			Bold(true)

	authValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	authErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	authSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)
)

func NewAuthLoginTUI(db *database.DB, salt []byte, masterPass string) *authLoginModel {
	providers := auth.GetAllProviders()

	return &authLoginModel{
		db:               db,
		salt:             salt,
		masterPass:       masterPass,
		providers:        providers,
		cursor:           0,
		step:             0,
		credentialInputs: make(map[string]string),
		width:            80,
		height:           24,
	}
}

func (m authLoginModel) Init() tea.Cmd {
	return nil
}

func (m authLoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == 0 || m.step == 3 {
				return m, tea.Quit
			}
			// Go back on other steps
			if m.step > 0 && m.step < 3 {
				m.step--
				m.errorMsg = ""
				if m.step == 0 {
					m.credentialInputs = make(map[string]string)
					m.currentField = 0
				}
				return m, nil
			}

		case "esc":
			if m.step > 0 && m.step < 3 {
				m.step--
				m.errorMsg = ""
				if m.step == 0 {
					m.credentialInputs = make(map[string]string)
					m.currentField = 0
				}
			} else {
				return m, tea.Quit
			}

		case "up", "k":
			if m.step == 0 && m.cursor > 0 {
				m.cursor--
			} else if m.step == 1 && m.currentField > 0 {
				m.currentField--
			}

		case "down", "j":
			if m.step == 0 && m.cursor < len(m.providers)-1 {
				m.cursor++
			} else if m.step == 1 {
				fields := m.selectedProvider.GetCredentialFields()
				if m.currentField < len(fields)-1 {
					m.currentField++
				}
			}

		case "tab":
			if m.step == 1 {
				fields := m.selectedProvider.GetCredentialFields()
				m.currentField = (m.currentField + 1) % len(fields)
			}

		case "enter":
			switch m.step {
			case 0: // Select provider
				m.selectedProvider = m.providers[m.cursor]
				m.step = 1
				m.errorMsg = ""
				// Initialize credential inputs
				for _, field := range m.selectedProvider.GetCredentialFields() {
					m.credentialInputs[field.Name] = ""
				}

			case 1: // Credentials entered
				// Validate required fields
				fields := m.selectedProvider.GetCredentialFields()
				for _, field := range fields {
					if field.Required && m.credentialInputs[field.Name] == "" {
						m.errorMsg = fmt.Sprintf("%s is required", field.Label)
						return m, nil
					}
				}

				// Attempt login
				if err := m.selectedProvider.Login(m.credentialInputs); err != nil {
					m.errorMsg = fmt.Sprintf("Login failed: %v", err)
					return m, nil
				}

				m.step = 2
				return m, m.performSync()

			case 3: // Done
				return m, tea.Quit
			}

		case "backspace":
			if m.step == 1 {
				fields := m.selectedProvider.GetCredentialFields()
				if m.currentField < len(fields) {
					fieldName := fields[m.currentField].Name
					if len(m.credentialInputs[fieldName]) > 0 {
						m.credentialInputs[fieldName] = m.credentialInputs[fieldName][:len(m.credentialInputs[fieldName])-1]
					}
				}
			}

		default:
			if m.step == 1 && len(msg.String()) == 1 {
				fields := m.selectedProvider.GetCredentialFields()
				if m.currentField < len(fields) {
					fieldName := fields[m.currentField].Name
					m.credentialInputs[fieldName] += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m *authLoginModel) performSync() tea.Cmd {
	return func() tea.Msg {
		// Sync passwords from provider
		passwords, err := m.selectedProvider.SyncPasswords()
		if err != nil {
			m.errorMsg = fmt.Sprintf("Sync failed: %v", err)
			m.step = 1
			return nil
		}

		if len(passwords) == 0 {
			m.errorMsg = "No passwords found"
			m.step = 1
			return nil
		}

		// Encrypt and save passwords
		encryptor := crypto.NewEncryptor(m.masterPass, m.salt)
		successCount := 0

		for _, pwd := range passwords {
			// Encrypt all fields
			if pwd.Name != "" {
				encrypted, err := encryptor.Encrypt(pwd.Name)
				if err != nil {
					continue
				}
				pwd.Name = encrypted
			}

			if pwd.Username != "" {
				encrypted, err := encryptor.Encrypt(pwd.Username)
				if err != nil {
					continue
				}
				pwd.Username = encrypted
			}

			if pwd.Password != "" {
				encrypted, err := encryptor.Encrypt(pwd.Password)
				if err != nil {
					continue
				}
				pwd.Password = encrypted
			}

			if pwd.URL != "" {
				encrypted, err := encryptor.Encrypt(pwd.URL)
				if err != nil {
					continue
				}
				pwd.URL = encrypted
			}

			if pwd.Notes != "" {
				encrypted, err := encryptor.Encrypt(pwd.Notes)
				if err != nil {
					continue
				}
				pwd.Notes = encrypted
			}

			// Encrypt custom fields
			for key, val := range pwd.Fields {
				encrypted, err := encryptor.Encrypt(val)
				if err != nil {
					continue
				}
				pwd.Fields[key] = encrypted
			}

			// Save to database
			if err := m.db.AddPassword(pwd); err != nil {
				continue
			}

			successCount++
		}

		m.syncedCount = successCount
		m.successMsg = fmt.Sprintf("Successfully synced %d passwords from %s", successCount, m.selectedProvider.GetName())
		m.step = 3
		return nil
	}
}

func (m authLoginModel) View() string {
	var s strings.Builder

	s.WriteString(authTitleStyle.Render("Auth Login"))
	s.WriteString("\n\n")

	switch m.step {
	case 0: // Select provider
		s.WriteString(authLabelStyle.Render("Select a provider to connect with:"))
		s.WriteString("\n\n")

		for i, provider := range m.providers {
			line := fmt.Sprintf("%-20s %s", provider.GetName(), provider.GetDescription())
			if i == m.cursor {
				s.WriteString(authSelectedStyle.Render("→ " + line))
			} else {
				s.WriteString(authNormalStyle.Render("  " + line))
			}
			s.WriteString("\n")
		}

		s.WriteString("\n")
		s.WriteString(authNormalStyle.Render("↑/↓: navigate • enter: select • q: quit"))

	case 1: // Enter credentials
		s.WriteString(authLabelStyle.Render(fmt.Sprintf("Connect to: %s", m.selectedProvider.GetName())))
		s.WriteString("\n\n")
		s.WriteString(authNormalStyle.Render(m.selectedProvider.GetDescription()))
		s.WriteString("\n\n")

		fields := m.selectedProvider.GetCredentialFields()
		for i, field := range fields {
			if i == m.currentField {
				s.WriteString(authSelectedStyle.Render("→ "))
			} else {
				s.WriteString(authNormalStyle.Render("  "))
			}

			s.WriteString(authLabelStyle.Render(field.Label + ": "))

			value := m.credentialInputs[field.Name]
			if field.Type == "password" && value != "" {
				s.WriteString(authValueStyle.Render(strings.Repeat("•", len(value))))
			} else {
				s.WriteString(authValueStyle.Render(value))
			}

			if i == m.currentField {
				s.WriteString(authValueStyle.Render("▋"))
			}

			if field.Placeholder != "" && value == "" {
				s.WriteString(" ")
				s.WriteString(authNormalStyle.Render("(" + field.Placeholder + ")"))
			}

			s.WriteString("\n")
		}

		if m.errorMsg != "" {
			s.WriteString("\n")
			s.WriteString(authErrorStyle.Render("✗ " + m.errorMsg))
			s.WriteString("\n")
		}

		s.WriteString("\n")
		s.WriteString(authNormalStyle.Render("↑/↓/tab: navigate fields • enter: connect • esc: back • q: quit"))

	case 2: // Syncing
		s.WriteString(authLabelStyle.Render("Syncing passwords..."))
		s.WriteString("\n\n")
		s.WriteString(authNormalStyle.Render(fmt.Sprintf("Connecting to %s and syncing your passwords.", m.selectedProvider.GetName())))
		s.WriteString("\n")
		s.WriteString(authNormalStyle.Render("Please wait..."))

	case 3: // Done
		if m.successMsg != "" {
			s.WriteString(authSuccessStyle.Render("✓ " + m.successMsg))
			s.WriteString("\n\n")
			s.WriteString(authNormalStyle.Render("Your passwords have been securely synced and encrypted."))
		} else if m.errorMsg != "" {
			s.WriteString(authErrorStyle.Render("✗ " + m.errorMsg))
		}
		s.WriteString("\n\n")
		s.WriteString(authNormalStyle.Render("Press any key to exit"))
	}

	return s.String()
}

func RunAuthLoginTUI(db *database.DB, salt []byte, masterPass string) error {
	p := tea.NewProgram(NewAuthLoginTUI(db, salt, masterPass))
	_, err := p.Run()
	return err
}
