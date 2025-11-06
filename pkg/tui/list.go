package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/models"
)

type listModel struct {
	db                *database.DB
	encryptor         *crypto.Encryptor
	passwords         []*models.Password
	filteredPasswords []*models.Password
	cursor            int
	searchInput       string
	showDetails       bool
	selectedPass      *models.Password
	showPassword      bool
	detailCursor      int
	detailFields      []detailField
	copiedMessage     string
	width             int
	height            int
}

type detailField struct {
	label string
	value string
}

var (
	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFAF00")).
			MarginBottom(1)

	listSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	listNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8"))

	listLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAF00")).
			Bold(true)

	listValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	listMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	listMetaDotStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Bold(true)

	listHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00")).
				Background(lipgloss.Color("#5F5F00")).
				Bold(true)

	listMetaHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00")).
				Background(lipgloss.Color("#3A3A00")).
				Bold(true).
				Italic(true)
)

func NewListTUI(db *database.DB, salt []byte, passphrase string) *listModel {
	encryptor := crypto.NewEncryptor(passphrase, salt)
	passwords, _ := db.ListPasswords()

	return &listModel{
		db:                db,
		encryptor:         encryptor,
		passwords:         passwords,
		filteredPasswords: passwords,
		cursor:            0,
		searchInput:       "",
		width:             80,
		height:            24,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.showDetails {
				m.showDetails = false
				m.showPassword = false
				return m, nil
			}
			return m, tea.Quit

		case "esc":
			if m.showDetails {
				m.showDetails = false
				m.showPassword = false
			} else if m.searchInput != "" {
				m.searchInput = ""
				m.filteredPasswords = m.passwords
				m.cursor = 0
			} else {
				return m, tea.Quit
			}

		case "tab":
			if m.showDetails {
				m.showPassword = !m.showPassword
			}

		case "up", "k":
			if m.showDetails {
				if m.detailCursor > 0 {
					m.detailCursor--
					m.copiedMessage = ""
				}
			} else if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.showDetails {
				if m.detailCursor < len(m.detailFields)-1 {
					m.detailCursor++
					m.copiedMessage = ""
				}
			} else if m.cursor < len(m.filteredPasswords)-1 {
				m.cursor++
			}

		case "c":
			if m.showDetails {
				allText := m.buildAllFieldsText()
				if err := copyToClipboard(allText); err == nil {
					m.copiedMessage = "✓ Copied all fields to clipboard"
				} else {
					m.copiedMessage = fmt.Sprintf("✗ Failed to copy: %v", err)
				}
			}

		case "enter":
			if m.showDetails {
				if len(m.detailFields) > 0 && m.detailCursor < len(m.detailFields) {
					field := m.detailFields[m.detailCursor]
					if err := copyToClipboard(field.value); err == nil {
						m.copiedMessage = fmt.Sprintf("✓ Copied %s to clipboard", field.label)
					} else {
						m.copiedMessage = fmt.Sprintf("✗ Failed to copy: %v", err)
					}
				}
			} else if len(m.filteredPasswords) > 0 && m.cursor < len(m.filteredPasswords) {
				m.selectedPass = m.filteredPasswords[m.cursor]
				m.showDetails = true
				m.showPassword = false
				m.detailCursor = 0
				m.copiedMessage = ""
				m.buildDetailFields()
			}

		case "backspace":
			if !m.showDetails && len(m.searchInput) > 0 {
				m.searchInput = m.searchInput[:len(m.searchInput)-1]
				m.filterPasswords()
				m.cursor = 0
			}

		default:
			if !m.showDetails && len(msg.String()) == 1 {
				m.searchInput += msg.String()
				m.filterPasswords()
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m *listModel) filterPasswords() {
	if m.searchInput == "" {
		m.filteredPasswords = m.passwords
		return
	}

	query := strings.ToLower(m.searchInput)
	filtered := []*models.Password{}

	for _, p := range m.passwords {
		// Decrypt all searchable fields
		decryptedName := p.Name
		if name, err := m.encryptor.Decrypt(p.Name); err == nil {
			decryptedName = name
		}

		decryptedUsername := ""
		if p.Username != "" {
			if username, err := m.encryptor.Decrypt(p.Username); err == nil {
				decryptedUsername = username
			}
		}

		decryptedURL := ""
		if p.URL != "" {
			if url, err := m.encryptor.Decrypt(p.URL); err == nil {
				decryptedURL = url
			}
		}

		decryptedNotes := ""
		if p.Notes != "" {
			if notes, err := m.encryptor.Decrypt(p.Notes); err == nil {
				decryptedNotes = notes
			}
		}

		// Search in all fields
		if strings.Contains(strings.ToLower(decryptedName), query) ||
			strings.Contains(strings.ToLower(decryptedUsername), query) ||
			strings.Contains(strings.ToLower(decryptedURL), query) ||
			strings.Contains(strings.ToLower(decryptedNotes), query) {
			filtered = append(filtered, p)
		}
	}

	m.filteredPasswords = filtered
}

func (m listModel) View() string {
	if m.showDetails {
		return m.renderDetails()
	}

	var s strings.Builder

	s.WriteString(listTitleStyle.Render("Password List"))
	s.WriteString("\n\n")

	if m.searchInput != "" {
		s.WriteString(listLabelStyle.Render("Search: "))
		s.WriteString(listValueStyle.Render(m.searchInput + "▋"))
		s.WriteString("\n\n")
	}

	if len(m.filteredPasswords) == 0 {
		s.WriteString(listNormalStyle.Render("No passwords found"))
	} else {
		s.WriteString(listNormalStyle.Render(fmt.Sprintf("Found %d password(s)", len(m.filteredPasswords))))
		s.WriteString("\n\n")

		for i, p := range m.filteredPasswords {
			// Decrypt name
			decryptedName := p.Name
			if name, err := m.encryptor.Decrypt(p.Name); err == nil {
				decryptedName = name
			}

			// Decrypt username
			decryptedUsername := ""
			if p.Username != "" {
				if username, err := m.encryptor.Decrypt(p.Username); err == nil {
					decryptedUsername = username
				}
			}

			// Decrypt URL
			decryptedURL := ""
			if p.URL != "" {
				if url, err := m.encryptor.Decrypt(p.URL); err == nil {
					decryptedURL = url
				}
			}

			// Decrypt notes (get first line only)
			decryptedNotes := ""
			if p.Notes != "" {
				if notes, err := m.encryptor.Decrypt(p.Notes); err == nil {
					// Take first line or first 30 chars
					lines := strings.Split(notes, "\n")
					if len(lines) > 0 {
						decryptedNotes = lines[0]
					}
				}
			}

			// Render the password entry
			cursor := "  "
			if i == m.cursor {
				cursor = listSelectedStyle.Render("→ ")
			}

			// First line: Name (with highlighting if searching)
			nameStyle := listNormalStyle
			nameHighlightStyle := listHighlightStyle
			if i == m.cursor {
				nameStyle = listSelectedStyle
			}
			s.WriteString(cursor)
			if m.searchInput != "" {
				s.WriteString(highlightText(decryptedName, m.searchInput, nameStyle, nameHighlightStyle))
			} else {
				s.WriteString(nameStyle.Render(decryptedName))
			}
			s.WriteString("\n")

			// Second line: metadata (username • website • notes)
			var metaParts []struct {
				text      string
				truncated string
			}
			if decryptedUsername != "" {
				metaParts = append(metaParts, struct {
					text      string
					truncated string
				}{decryptedUsername, truncateString(decryptedUsername, 25)})
			}
			if decryptedURL != "" {
				metaParts = append(metaParts, struct {
					text      string
					truncated string
				}{decryptedURL, truncateString(decryptedURL, 30)})
			}
			if decryptedNotes != "" {
				metaParts = append(metaParts, struct {
					text      string
					truncated string
				}{decryptedNotes, truncateString(decryptedNotes, 30)})
			}

			if len(metaParts) > 0 {
				s.WriteString("  ") // Indent
				for idx, part := range metaParts {
					// Apply highlighting to metadata if searching
					if m.searchInput != "" {
						s.WriteString(highlightText(part.truncated, m.searchInput, listMetaStyle, listMetaHighlightStyle))
					} else {
						s.WriteString(listMetaStyle.Render(part.truncated))
					}
					if idx < len(metaParts)-1 {
						s.WriteString(" ")
						s.WriteString(listMetaDotStyle.Render("•"))
						s.WriteString(" ")
					}
				}
				s.WriteString("\n")
			}

			// Add spacing between entries
			if i < len(m.filteredPasswords)-1 {
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(listNormalStyle.Render("↑/↓: navigate • enter: view details • type to search • esc: clear/back • q: quit"))

	return s.String()
}

func (m listModel) renderDetails() string {
	if m.selectedPass == nil {
		return "No password selected"
	}

	var s strings.Builder

	s.WriteString(listTitleStyle.Render("Password Details"))
	s.WriteString("\n\n")

	for i, field := range m.detailFields {
		isPasswordField := field.label == "Password" || field.label == "Cvv" || field.label == "Number"

		if i == m.detailCursor {
			s.WriteString(listSelectedStyle.Render("→ "))
		} else {
			s.WriteString(listNormalStyle.Render("  "))
		}

		s.WriteString(listLabelStyle.Render(field.label + ": "))

		if isPasswordField && !m.showPassword {
			s.WriteString(listValueStyle.Render(strings.Repeat("•", len(field.value))))
		} else {
			s.WriteString(listValueStyle.Render(field.value))
		}

		if isPasswordField {
			s.WriteString(" ")
			if m.showPassword {
				s.WriteString(listLabelStyle.Render("👁"))
			} else {
				s.WriteString(listLabelStyle.Render("👁‍🗨"))
			}
		}

		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(listLabelStyle.Render("Created: "))
	s.WriteString(listNormalStyle.Render(m.selectedPass.CreatedAt.Format("2006-01-02 15:04:05")))
	s.WriteString("\n")

	s.WriteString(listLabelStyle.Render("Updated: "))
	s.WriteString(listNormalStyle.Render(m.selectedPass.UpdatedAt.Format("2006-01-02 15:04:05")))
	s.WriteString("\n\n")

	if m.copiedMessage != "" {
		if strings.HasPrefix(m.copiedMessage, "✓") {
			s.WriteString(addSuccessStyle.Render(m.copiedMessage))
		} else {
			s.WriteString(addErrorStyle.Render(m.copiedMessage))
		}
		s.WriteString("\n\n")
	}

	s.WriteString(listLabelStyle.Render("Type: "))
	s.WriteString(listValueStyle.Render(string(m.selectedPass.Type)))
	s.WriteString("\n")

	s.WriteString(listNormalStyle.Render("↑/↓: select field • enter: copy • c: copy all • tab: toggle password • q/esc: go back"))

	return s.String()
}

func (m *listModel) buildDetailFields() {
	m.detailFields = []detailField{}

	if m.selectedPass == nil {
		return
	}

	decryptedName := m.selectedPass.Name
	if name, err := m.encryptor.Decrypt(m.selectedPass.Name); err == nil {
		decryptedName = name
	}
	m.detailFields = append(m.detailFields, detailField{"Name", decryptedName})

	if m.selectedPass.Username != "" {
		decryptedUsername := m.selectedPass.Username
		if username, err := m.encryptor.Decrypt(m.selectedPass.Username); err == nil {
			decryptedUsername = username
		}
		m.detailFields = append(m.detailFields, detailField{"Username", decryptedUsername})
	}

	if m.selectedPass.Password != "" {
		decryptedPassword := m.selectedPass.Password
		if password, err := m.encryptor.Decrypt(m.selectedPass.Password); err == nil {
			decryptedPassword = password
		}
		m.detailFields = append(m.detailFields, detailField{"Password", decryptedPassword})
	}

	if m.selectedPass.URL != "" {
		decryptedURL := m.selectedPass.URL
		if url, err := m.encryptor.Decrypt(m.selectedPass.URL); err == nil {
			decryptedURL = url
		}
		m.detailFields = append(m.detailFields, detailField{"URL", decryptedURL})
	}

	for key, val := range m.selectedPass.Fields {
		decryptedVal := val
		if decrypted, err := m.encryptor.Decrypt(val); err == nil {
			decryptedVal = decrypted
		}
		label := strings.Title(strings.ReplaceAll(key, "_", " "))
		m.detailFields = append(m.detailFields, detailField{label, decryptedVal})
	}

	if m.selectedPass.Notes != "" {
		decryptedNotes := m.selectedPass.Notes
		if notes, err := m.encryptor.Decrypt(m.selectedPass.Notes); err == nil {
			decryptedNotes = notes
		}
		m.detailFields = append(m.detailFields, detailField{"Notes", decryptedNotes})
	}
}

func (m *listModel) buildAllFieldsText() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Type: %s\n", m.selectedPass.Type))

	for _, field := range m.detailFields {
		s.WriteString(fmt.Sprintf("%s: %s\n", field.label, field.value))
	}

	s.WriteString(fmt.Sprintf("\nCreated: %s\n", m.selectedPass.CreatedAt.Format("2006-01-02 15:04:05")))
	s.WriteString(fmt.Sprintf("Updated: %s\n", m.selectedPass.UpdatedAt.Format("2006-01-02 15:04:05")))

	return s.String()
}

func copyToClipboard(text string) error {
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

func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

// highlightText highlights matching query in text with the given style
func highlightText(text, query string, normalStyle, highlightStyle lipgloss.Style) string {
	if query == "" {
		return normalStyle.Render(text)
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	if !strings.Contains(lowerText, lowerQuery) {
		return normalStyle.Render(text)
	}

	var result strings.Builder
	remaining := text
	lowerRemaining := lowerText

	for {
		index := strings.Index(lowerRemaining, lowerQuery)
		if index == -1 {
			// No more matches, add remaining text
			if len(remaining) > 0 {
				result.WriteString(normalStyle.Render(remaining))
			}
			break
		}

		// Add text before match
		if index > 0 {
			result.WriteString(normalStyle.Render(remaining[:index]))
		}

		// Add highlighted match
		result.WriteString(highlightStyle.Render(remaining[index : index+len(query)]))

		// Move to text after match
		remaining = remaining[index+len(query):]
		lowerRemaining = lowerRemaining[index+len(query):]
	}

	return result.String()
}

func RunListTUI(db *database.DB, salt []byte, passphrase string) error {
	p := tea.NewProgram(
		NewListTUI(db, salt, passphrase),
	)
	_, err := p.Run()
	return err
}
