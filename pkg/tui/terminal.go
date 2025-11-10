package tui

import (
	"fmt"
	"strings"
)

const (
	termColorReset  = "\033[0m"
	termColorGrey   = "\033[38;5;243m"
	termColorOrange = "\033[38;5;214m"
	termColorWhite  = "\033[38;5;255m"
	termColorBold   = "\033[1m"
)

// PromptPassword displays a styled password prompt with bullet feedback
func PromptPassword(message string, optional bool) (string, error) {
	// Build the prompt with styling
	var prompt strings.Builder
	prompt.WriteString(termColorOrange)
	prompt.WriteString(termColorBold)
	prompt.WriteString("→ ")
	prompt.WriteString(termColorReset)
	prompt.WriteString(termColorWhite)
	prompt.WriteString(message)
	if optional {
		prompt.WriteString(" ")
		prompt.WriteString(termColorGrey)
		prompt.WriteString("(or press Enter for none)")
		prompt.WriteString(termColorReset)
	}
	prompt.WriteString(termColorWhite)
	prompt.WriteString(": ")
	prompt.WriteString(termColorReset)

	fmt.Print(prompt.String())

	password, err := readPasswordWithBullets(prompt.String(), true)
	if err != nil {
		return "", err
	}

	fmt.Println() // New line after password entry
	return password, nil
}
