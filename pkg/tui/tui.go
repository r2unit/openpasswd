package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/models"
)

type App struct {
	db        *database.DB
	salt      []byte
	encryptor *crypto.Encryptor
	scanner   *bufio.Scanner
}

func New(db *database.DB, salt []byte) *App {
	return &App{
		db:      db,
		salt:    salt,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (a *App) Run() error {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              OpenPasswd - Password Manager                ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝\n")

	passphrase, err := a.getPassphrase()
	if err != nil {
		return err
	}

	a.encryptor = crypto.NewEncryptor(passphrase, a.salt)

	for {
		a.showMenu()
		choice := a.readInput("Select an option: ")

		switch choice {
		case "1":
			a.listPasswords()
		case "2":
			a.addPassword()
		case "3":
			a.viewPassword()
		case "4":
			a.searchPasswords()
		case "5":
			a.updatePassword()
		case "6":
			a.deletePassword()
		case "7", "q", "quit", "exit":
			fmt.Println("\nGoodbye!\n")
			return nil
		default:
			fmt.Println("Invalid option. Please try again.\n")
		}
	}
}

func (a *App) getPassphrase() (string, error) {
	fmt.Print("Enter master passphrase: ")
	password, err := readPassword()
	if err != nil {
		return "", err
	}
	fmt.Println()
	return password, nil
}

func (a *App) showMenu() {
	fmt.Println("\n┌───────────────── Menu ─────────────────┐")
	fmt.Println("│  1. List all passwords                 │")
	fmt.Println("│  2. Add new password                   │")
	fmt.Println("│  3. View password                      │")
	fmt.Println("│  4. Search passwords                   │")
	fmt.Println("│  5. Update password                    │")
	fmt.Println("│  6. Delete password                    │")
	fmt.Println("│  7. Exit                               │")
	fmt.Println("└────────────────────────────────────────┘\n")
}

func (a *App) readInput(prompt string) string {
	fmt.Print(prompt)
	a.scanner.Scan()
	return strings.TrimSpace(a.scanner.Text())
}

func (a *App) readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := readPassword()
	if err != nil {
		return "", err
	}
	fmt.Println()
	return password, nil
}

func (a *App) listPasswords() {
	passwords, err := a.db.ListPasswords()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(passwords) == 0 {
		fmt.Println("\nNo passwords stored yet.\n")
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Stored Passwords                      ║")
	fmt.Println("╠═══╦═══════════════════════════╦═══════════════════════════╣")
	fmt.Println("║ ID║ Name                      ║ Username                  ║")
	fmt.Println("╠═══╬═══════════════════════════╬═══════════════════════════╣")

	for _, p := range passwords {
		fmt.Printf("║%-3d║ %-25s ║ %-25s ║\n", p.ID, truncate(p.Name, 25), truncate(p.Username, 25))
	}

	fmt.Println("╚═══╩═══════════════════════════╩═══════════════════════════╝\n")
}

func (a *App) addPassword() {
	fmt.Println("\n┌─────── Add New Password ───────┐")

	name := a.readInput("Name: ")
	username := a.readInput("Username: ")
	password, err := a.readPassword("Password: ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	url := a.readInput("URL (optional): ")
	notes := a.readInput("Notes (optional): ")

	encrypted, err := a.encryptor.Encrypt(password)
	if err != nil {
		fmt.Printf("Error encrypting password: %v\n", err)
		return
	}

	p := &models.Password{
		Name:     name,
		Username: username,
		Password: encrypted,
		URL:      url,
		Notes:    notes,
	}

	if err := a.db.AddPassword(p); err != nil {
		fmt.Printf("Error saving password: %v\n", err)
		return
	}

	fmt.Println("Password added successfully!\n")
}

func (a *App) viewPassword() {
	idStr := a.readInput("Enter password ID: ")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Println("Invalid ID\n")
		return
	}

	p, err := a.db.GetPassword(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	decrypted, err := a.encryptor.Decrypt(p.Password)
	if err != nil {
		fmt.Printf("Error decrypting password: %v\n", err)
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Printf("║ ID:       %-48d║\n", p.ID)
	fmt.Printf("║ Name:     %-48s║\n", p.Name)
	fmt.Printf("║ Username: %-48s║\n", p.Username)
	fmt.Printf("║ Password: %-48s║\n", decrypted)
	fmt.Printf("║ URL:      %-48s║\n", p.URL)
	fmt.Printf("║ Notes:    %-48s║\n", p.Notes)
	fmt.Printf("║ Created:  %-48s║\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("║ Updated:  %-48s║\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("╚═══════════════════════════════════════════════════════════╝\n")
}

func (a *App) searchPasswords() {
	query := a.readInput("Search query: ")

	passwords, err := a.db.SearchPasswords(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(passwords) == 0 {
		fmt.Println("\nNo passwords found.\n")
		return
	}

	fmt.Printf("\nFound %d password(s):\n\n", len(passwords))
	fmt.Println("╔═══╦═══════════════════════════╦═══════════════════════════╗")
	fmt.Println("║ ID║ Name                      ║ Username                  ║")
	fmt.Println("╠═══╬═══════════════════════════╬═══════════════════════════╣")

	for _, p := range passwords {
		fmt.Printf("║%-3d║ %-25s ║ %-25s ║\n", p.ID, truncate(p.Name, 25), truncate(p.Username, 25))
	}

	fmt.Println("╚═══╩═══════════════════════════╩═══════════════════════════╝\n")
}

func (a *App) updatePassword() {
	idStr := a.readInput("Enter password ID to update: ")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Println("Invalid ID\n")
		return
	}

	p, err := a.db.GetPassword(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n┌─────── Update Password ───────┐")
	fmt.Println("(Leave blank to keep current value)")

	name := a.readInput(fmt.Sprintf("Name [%s]: ", p.Name))
	if name != "" {
		p.Name = name
	}

	username := a.readInput(fmt.Sprintf("Username [%s]: ", p.Username))
	if username != "" {
		p.Username = username
	}

	passwordInput, err := a.readPassword("New password (leave blank to keep current): ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if passwordInput != "" {
		encrypted, err := a.encryptor.Encrypt(passwordInput)
		if err != nil {
			fmt.Printf("Error encrypting password: %v\n", err)
			return
		}
		p.Password = encrypted
	}

	url := a.readInput(fmt.Sprintf("URL [%s]: ", p.URL))
	if url != "" {
		p.URL = url
	}

	notes := a.readInput(fmt.Sprintf("Notes [%s]: ", p.Notes))
	if notes != "" {
		p.Notes = notes
	}

	if err := a.db.UpdatePassword(p); err != nil {
		fmt.Printf("Error updating password: %v\n", err)
		return
	}

	fmt.Println("Password updated successfully!\n")
}

func (a *App) deletePassword() {
	idStr := a.readInput("Enter password ID to delete: ")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Println("Invalid ID\n")
		return
	}

	confirm := a.readInput("Are you sure? (yes/no): ")
	if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
		fmt.Println("Cancelled\n")
		return
	}

	if err := a.db.DeletePassword(id); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Password deleted successfully!\n")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
