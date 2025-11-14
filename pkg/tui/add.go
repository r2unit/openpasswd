package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/r2unit/openpasswd/pkg/config"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/models"
)

type addModel struct {
	db              *database.DB
	salt            []byte
	encryptor       *crypto.Encryptor
	passphrase      string
	step            int
	passwordType    string
	cursor          int
	inputs          map[string]string
	inputOrder      []string
	currentInput    string
	currentInputIdx int
	commandInput    string
	err             error
	success         bool
	loading         bool
	spinner         int
	width           int
	height          int
	showPassword    map[string]bool
	successTimer    int
	keybindings     config.Keybindings
}

var passwordTypes = []struct {
	name string
	desc string
}{
	{"login", "Username and password"},
	{"card", "Credit/debit card"},
	{"note", "Secure note"},
	{"identity", "Personal identity"},
	{"password", "Simple password"},
	{"other", "Other credential"},
}

var (
	_ = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#5FAFFF")).
		MarginBottom(1)

	addSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFAF00")).
				Bold(true)

	addNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A8A8A8"))

	addLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAF00")).
			Bold(true)

	addInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	addErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			Bold(true)

	addSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D75F")).
			Bold(true)

	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

func NewAddTUI(db *database.DB, salt []byte, passphrase string, passwordType string) *addModel {
	var encryptor *crypto.Encryptor
	if passphrase != "" {
		encryptor = crypto.NewEncryptor(passphrase, salt)
	}

	keybindings, _ := config.LoadKeybindings()

	m := &addModel{
		db:           db,
		salt:         salt,
		encryptor:    encryptor,
		passphrase:   passphrase,
		step:         0,
		inputs:       make(map[string]string),
		showPassword: make(map[string]bool),
		width:        80,
		height:       24,
		keybindings:  keybindings,
		commandInput: "",
	}

	if passwordType != "" {
		m.passwordType = passwordType
		m.step = 1
		m.setupInputs()
	}

	return m
}

func (m *addModel) setupInputs() {
	switch m.passwordType {
	case "login":
		m.inputOrder = []string{"name", "username", "password", "url", "notes"}
	case "card":
		m.inputOrder = []string{"name", "cardholder", "number", "expiry", "cvv", "notes"}
	case "note":
		m.inputOrder = []string{"name", "content"}
	case "identity":
		m.inputOrder = []string{"name", "full_name", "email", "phone", "address", "notes"}
	case "password":
		m.inputOrder = []string{"name", "password", "notes"}
	case "other":
		m.inputOrder = []string{"name", "value", "notes"}
	}

	if len(m.inputOrder) > 0 {
		m.currentInput = m.inputOrder[0]
		m.currentInputIdx = 0
	}
}

type tickMsg struct{}

type saveResultMsg struct {
	err     error
	success bool
}

func tick() tea.Cmd {
	return tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m addModel) Init() tea.Cmd {
	return tick()
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case saveResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else if msg.success {
			m.success = true
			m.successTimer = 15
			return m, tick()
		}
		return m, nil

	case tickMsg:
		if m.loading {
			m.spinner = (m.spinner + 1) % len(spinnerFrames)
			return m, tick()
		}
		if m.success && m.successTimer > 0 {
			m.successTimer--
			if m.successTimer == 0 {
				m.success = false
				m.step = 0
				m.passwordType = ""
				m.inputs = make(map[string]string)
				m.showPassword = make(map[string]bool)
				m.cursor = 0
				return m, nil
			}
			return m, tick()
		}
		return m, nil

	case tea.KeyMsg:
		if m.success {
			return m, tea.Quit
		}

		key := msg.String()

		// Handle command mode (nvim-style) - only when not in input mode
		if key == ":" && m.commandInput == "" && m.step == 0 {
			m.commandInput = ":"
			return m, nil
		}

		// If in command mode
		if m.commandInput != "" {
			if key == "enter" {
				// Execute command
				if m.commandInput == m.keybindings.Quit {
					return m, tea.Quit
				}
				m.commandInput = ""
				return m, nil
			} else if key == "backspace" {
				if len(m.commandInput) > 0 {
					m.commandInput = m.commandInput[:len(m.commandInput)-1]
				}
				return m, nil
			} else if key == "esc" {
				m.commandInput = ""
				return m, nil
			} else if len(key) == 1 {
				m.commandInput += key
				return m, nil
			}
			return m, nil
		}

		// Normal mode key handling
		switch key {
		case m.keybindings.QuitAlt:
			if m.step == 0 || m.err != nil {
				return m, tea.Quit
			}

		case m.keybindings.Back:
			if m.step == 1 {
				m.step = 0
				m.passwordType = ""
				m.inputs = make(map[string]string)
				m.showPassword = make(map[string]bool)
				return m, nil
			}

		case "tab":
			if m.step == 1 {
				key := m.inputOrder[m.currentInputIdx]
				if key == "password" || key == "cvv" {
					m.showPassword[key] = !m.showPassword[key]
				}
			}

		case m.keybindings.Up, m.keybindings.UpAlt:
			if m.step == 0 {
				if m.cursor > 0 {
					m.cursor--
				}
			} else if m.step == 1 {
				if m.currentInputIdx > 0 {
					m.currentInputIdx--
					m.currentInput = m.inputOrder[m.currentInputIdx]
				}
			}

		case m.keybindings.Down, m.keybindings.DownAlt:
			if m.step == 0 {
				if m.cursor < len(passwordTypes)-1 {
					m.cursor++
				}
			} else if m.step == 1 {
				if m.currentInputIdx < len(m.inputOrder)-1 {
					m.currentInputIdx++
					m.currentInput = m.inputOrder[m.currentInputIdx]
				}
			}

		case m.keybindings.Select:
			if m.step == 0 {
				m.passwordType = passwordTypes[m.cursor].name
				m.step = 1
				m.setupInputs()
			} else if m.step == 1 {
				idx := -1
				for i, key := range m.inputOrder {
					if key == m.currentInput {
						idx = i
						break
					}
				}

				if idx == len(m.inputOrder)-1 {
					m.loading = true
					return m, m.savePassword()
				} else if idx >= 0 {
					m.currentInput = m.inputOrder[idx+1]
				}
			}

		case "backspace":
			if m.step == 1 {
				if val, ok := m.inputs[m.currentInput]; ok && len(val) > 0 {
					m.inputs[m.currentInput] = val[:len(val)-1]
				}
			}

		default:
			if m.step == 1 && len(msg.String()) == 1 {
				m.inputs[m.currentInput] += msg.String()
			}
		}
	}

	return m, nil
}

func (m *addModel) savePassword() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs["name"]
		if name == "" {
			return saveResultMsg{err: fmt.Errorf("name is required")}
		}

		if m.encryptor == nil {
			m.encryptor = crypto.NewEncryptor(m.passphrase, m.salt)
		}

		password := &models.Password{
			Type:   models.PasswordType(m.passwordType),
			Fields: make(map[string]string),
		}

		encryptedName, err := m.encryptor.Encrypt(name)
		if err != nil {
			return saveResultMsg{err: err}
		}
		password.Name = encryptedName

		switch m.passwordType {
		case "login":
			if m.inputs["username"] != "" {
				encryptedUsername, err := m.encryptor.Encrypt(m.inputs["username"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Username = encryptedUsername
			}
			if m.inputs["password"] != "" {
				encryptedPassword, err := m.encryptor.Encrypt(m.inputs["password"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Password = encryptedPassword
			}
			if m.inputs["url"] != "" {
				encryptedURL, err := m.encryptor.Encrypt(m.inputs["url"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.URL = encryptedURL
			}
			if m.inputs["notes"] != "" {
				encryptedNotes, err := m.encryptor.Encrypt(m.inputs["notes"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Notes = encryptedNotes
			}
		case "card":
			for _, field := range []string{"cardholder", "number", "expiry", "cvv"} {
				if m.inputs[field] != "" {
					encrypted, err := m.encryptor.Encrypt(m.inputs[field])
					if err != nil {
						return saveResultMsg{err: err}
					}
					password.Fields[field] = encrypted
				}
			}
			if m.inputs["notes"] != "" {
				encryptedNotes, err := m.encryptor.Encrypt(m.inputs["notes"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Notes = encryptedNotes
			}
		case "note":
			if m.inputs["content"] != "" {
				encryptedContent, err := m.encryptor.Encrypt(m.inputs["content"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Notes = encryptedContent
			}
		case "identity":
			for _, field := range []string{"full_name", "email", "phone", "address"} {
				if m.inputs[field] != "" {
					encrypted, err := m.encryptor.Encrypt(m.inputs[field])
					if err != nil {
						return saveResultMsg{err: err}
					}
					password.Fields[field] = encrypted
				}
			}
			if m.inputs["notes"] != "" {
				encryptedNotes, err := m.encryptor.Encrypt(m.inputs["notes"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Notes = encryptedNotes
			}
		case "password":
			if m.inputs["password"] != "" {
				encryptedPassword, err := m.encryptor.Encrypt(m.inputs["password"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Password = encryptedPassword
			}
			if m.inputs["notes"] != "" {
				encryptedNotes, err := m.encryptor.Encrypt(m.inputs["notes"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Notes = encryptedNotes
			}
		case "other":
			if m.inputs["value"] != "" {
				encrypted, err := m.encryptor.Encrypt(m.inputs["value"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Fields["value"] = encrypted
			}
			if m.inputs["notes"] != "" {
				encryptedNotes, err := m.encryptor.Encrypt(m.inputs["notes"])
				if err != nil {
					return saveResultMsg{err: err}
				}
				password.Notes = encryptedNotes
			}
		}

		if err := m.db.AddPassword(password); err != nil {
			return saveResultMsg{err: err}
		}

		return saveResultMsg{success: true}
	}
}

func (m addModel) View() string {
	if m.success {
		return addSuccessStyle.Render("✓ ") + "Password saved successfully!\n"
	}

	if m.err != nil {
		return addErrorStyle.Render("✗ ") + fmt.Sprintf("Error: %v\n\nPress ':q' or 'ctrl+c' to exit", m.err)
	}

	if m.loading {
		return addSuccessStyle.Render(spinnerFrames[m.spinner]) + " Saving password...\n"
	}

	var s strings.Builder

	if m.step == 0 {
		// Tree-style header
		s.WriteString(addNormalStyle.Render("┌  "))
		s.WriteString(addLabelStyle.Render("Add credential"))
		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("│"))
		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("◆  "))
		s.WriteString(addLabelStyle.Render("Select type"))
		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("│"))
		s.WriteString("\n")

		for i, pt := range passwordTypes {
			prefix := "│  "
			bullet := "○"
			nameStyle := addNormalStyle

			if i == m.cursor {
				bullet = "●"
				nameStyle = addSelectedStyle
			}

			s.WriteString(addNormalStyle.Render(prefix))
			s.WriteString(nameStyle.Render(bullet + "  "))
			s.WriteString(nameStyle.Render(pt.name))
			s.WriteString(addNormalStyle.Render(" "))
			s.WriteString(addNormalStyle.Render("("))
			s.WriteString(addNormalStyle.Render(pt.desc))
			s.WriteString(addNormalStyle.Render(")"))
			s.WriteString("\n")
		}

		s.WriteString("\n")
		if m.commandInput != "" {
			s.WriteString(addSelectedStyle.Render(m.commandInput + "▋"))
		} else {
			s.WriteString(addNormalStyle.Render("↑/↓ or k/j: navigate • enter: select • :q or ctrl+c: quit"))
		}
	} else if m.step == 1 {
		// Tree-style header
		s.WriteString(addNormalStyle.Render("┌  "))
		s.WriteString(addLabelStyle.Render("Add credential"))
		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("│"))
		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("◆  "))
		s.WriteString(addLabelStyle.Render(strings.Title(m.passwordType)))
		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("│"))
		s.WriteString("\n")

		for idx, key := range m.inputOrder {
			label := strings.ReplaceAll(strings.Title(key), "_", " ")
			prefix := "│  "

			s.WriteString(addNormalStyle.Render(prefix))

			// Show field label
			if key == m.currentInput {
				s.WriteString(addSelectedStyle.Render(label + ": "))
			} else {
				s.WriteString(addLabelStyle.Render(label + ": "))
			}

			value := m.inputs[key]
			isPasswordField := key == "password" || key == "cvv"

			if key == m.currentInput {
				if isPasswordField {
					if m.showPassword[key] {
						s.WriteString(addInputStyle.Render(value + "▋"))
						s.WriteString(" ")
						s.WriteString(addLabelStyle.Render("(visible)"))
					} else {
						s.WriteString(addInputStyle.Render(strings.Repeat("•", len(value)) + "▋"))
						s.WriteString(" ")
						s.WriteString(addNormalStyle.Render("(hidden)"))
					}
				} else {
					s.WriteString(addInputStyle.Render(value + "▋"))
				}
			} else {
				if isPasswordField {
					if m.showPassword[key] {
						s.WriteString(addNormalStyle.Render(value))
					} else {
						s.WriteString(addNormalStyle.Render(strings.Repeat("•", len(value))))
					}
				} else {
					s.WriteString(addNormalStyle.Render(value))
				}
			}
			s.WriteString("\n")

			// Add separator between fields
			if idx < len(m.inputOrder)-1 {
				s.WriteString(addNormalStyle.Render("│"))
				s.WriteString("\n")
			}
		}

		s.WriteString("\n")
		s.WriteString(addNormalStyle.Render("↑/↓ or k/j: navigate • tab: toggle password • enter: next/save • esc: back"))
	}

	return s.String()
}

func RunAddTUI(db *database.DB, salt []byte, passphrase string, passwordType string) error {
	p := tea.NewProgram(
		NewAddTUI(db, salt, passphrase, passwordType),
	)
	_, err := p.Run()
	return err
}
