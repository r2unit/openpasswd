package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/sources"
)

type importModel struct {
	db             *database.DB
	salt           []byte
	masterPass     string
	importers      []sources.Importer
	cursor         int
	step           int // 0: select source, 1: enter file path, 2: enter passphrase (if needed), 3: importing, 4: done
	selectedSource sources.Importer
	filePath       string
	filePassphrase string
	filePassInput  string
	errorMsg       string
	successMsg     string
	importedCount  int
	width          int
	height         int
}

var (
	importTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFAF00")).
				MarginBottom(1)

	importSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	importNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A8A8A8"))

	importLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	importValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	importErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	importSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)
)

func NewImportTUI(db *database.DB, salt []byte, masterPass string) *importModel {
	importers := sources.GetAvailableImporters()

	return &importModel{
		db:         db,
		salt:       salt,
		masterPass: masterPass,
		importers:  importers,
		cursor:     0,
		step:       0,
		width:      80,
		height:     24,
	}
}

func (m importModel) Init() tea.Cmd {
	return nil
}

func (m importModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == 0 || m.step == 4 {
				return m, tea.Quit
			}
			// Go back on other steps
			if m.step > 0 && m.step < 4 {
				m.step--
				m.errorMsg = ""
				if m.step == 0 {
					m.filePath = ""
					m.filePassphrase = ""
					m.filePassInput = ""
				}
				return m, nil
			}

		case "esc":
			if m.step > 0 && m.step < 4 {
				m.step--
				m.errorMsg = ""
				if m.step == 0 {
					m.filePath = ""
					m.filePassphrase = ""
					m.filePassInput = ""
				}
			} else {
				return m, tea.Quit
			}

		case "up", "k":
			if m.step == 0 && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.step == 0 && m.cursor < len(m.importers)-1 {
				m.cursor++
			}

		case "enter":
			switch m.step {
			case 0: // Select source
				m.selectedSource = m.importers[m.cursor]
				m.step = 1
				m.errorMsg = ""
			case 1: // File path entered
				if m.filePath == "" {
					m.errorMsg = "File path cannot be empty"
					return m, nil
				}
				// Check if file exists
				if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
					m.errorMsg = "File does not exist: " + m.filePath
					return m, nil
				}
				// Check if source needs passphrase (for encrypted files)
				ext := strings.ToLower(m.filePath[strings.LastIndex(m.filePath, "."):])
				if ext == ".zip" || ext == ".pgp" {
					m.step = 2 // Ask for passphrase
				} else {
					m.step = 3 // Start import
					return m, m.performImport()
				}
			case 2: // Passphrase entered
				m.filePassphrase = m.filePassInput
				m.step = 3
				return m, m.performImport()
			case 4: // Done
				return m, tea.Quit
			}

		case "backspace":
			if m.step == 1 && len(m.filePath) > 0 {
				m.filePath = m.filePath[:len(m.filePath)-1]
			} else if m.step == 2 && len(m.filePassInput) > 0 {
				m.filePassInput = m.filePassInput[:len(m.filePassInput)-1]
			}

		default:
			if m.step == 1 && len(msg.String()) == 1 {
				m.filePath += msg.String()
			} else if m.step == 2 && len(msg.String()) == 1 {
				m.filePassInput += msg.String()
			}
		}
	}

	return m, nil
}

func (m *importModel) performImport() tea.Cmd {
	return func() tea.Msg {
		passwords, err := m.selectedSource.Import(m.filePath, m.filePassphrase)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Import failed: %v", err)
			m.step = 1
			return nil
		}

		if len(passwords) == 0 {
			m.errorMsg = "No passwords found in file"
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

		m.importedCount = successCount
		m.successMsg = fmt.Sprintf("Successfully imported %d passwords", successCount)
		m.step = 4
		return nil
	}
}

func (m importModel) View() string {
	var s strings.Builder

	s.WriteString(importTitleStyle.Render("Import Passwords"))
	s.WriteString("\n\n")

	switch m.step {
	case 0: // Select source
		s.WriteString(importLabelStyle.Render("Select password manager to import from:"))
		s.WriteString("\n\n")

		for i, importer := range m.importers {
			line := fmt.Sprintf("%-15s %s", importer.GetName(), importer.GetDescription())
			if i == m.cursor {
				s.WriteString(importSelectedStyle.Render("→ " + line))
			} else {
				s.WriteString(importNormalStyle.Render("  " + line))
			}
			s.WriteString("\n")
		}

		s.WriteString("\n")
		s.WriteString(importNormalStyle.Render("↑/↓: navigate • enter: select • q: quit"))

	case 1: // Enter file path
		s.WriteString(importLabelStyle.Render(fmt.Sprintf("Import from: %s", m.selectedSource.GetName())))
		s.WriteString("\n\n")
		s.WriteString(importLabelStyle.Render("Enter file path: "))
		s.WriteString(importValueStyle.Render(m.filePath + "▋"))
		s.WriteString("\n\n")

		if m.errorMsg != "" {
			s.WriteString(importErrorStyle.Render("✗ " + m.errorMsg))
			s.WriteString("\n\n")
		}

		s.WriteString(importNormalStyle.Render("Type the full path to your export file"))
		s.WriteString("\n")
		s.WriteString(importNormalStyle.Render("enter: continue • esc: back • q: quit"))

	case 2: // Enter passphrase
		s.WriteString(importLabelStyle.Render(fmt.Sprintf("Import from: %s", m.selectedSource.GetName())))
		s.WriteString("\n\n")
		s.WriteString(importLabelStyle.Render("File: "))
		s.WriteString(importValueStyle.Render(m.filePath))
		s.WriteString("\n\n")
		s.WriteString(importLabelStyle.Render("Enter export passphrase: "))
		s.WriteString(importValueStyle.Render(strings.Repeat("•", len(m.filePassInput)) + "▋"))
		s.WriteString("\n\n")

		if m.errorMsg != "" {
			s.WriteString(importErrorStyle.Render("✗ " + m.errorMsg))
			s.WriteString("\n\n")
		}

		s.WriteString(importNormalStyle.Render("Enter the passphrase you used when exporting"))
		s.WriteString("\n")
		s.WriteString(importNormalStyle.Render("enter: import • esc: back • q: quit"))

	case 3: // Importing
		s.WriteString(importLabelStyle.Render("Importing passwords..."))
		s.WriteString("\n\n")
		s.WriteString(importNormalStyle.Render("Please wait while your passwords are being imported and encrypted."))

	case 4: // Done
		if m.successMsg != "" {
			s.WriteString(importSuccessStyle.Render("✓ " + m.successMsg))
			s.WriteString("\n\n")
			s.WriteString(importNormalStyle.Render("Your passwords have been securely imported and encrypted."))
		} else if m.errorMsg != "" {
			s.WriteString(importErrorStyle.Render("✗ " + m.errorMsg))
		}
		s.WriteString("\n\n")
		s.WriteString(importNormalStyle.Render("Press any key to exit"))
	}

	return s.String()
}

func RunImportTUI(db *database.DB, salt []byte, masterPass string) error {
	p := tea.NewProgram(NewImportTUI(db, salt, masterPass))
	_, err := p.Run()
	return err
}
