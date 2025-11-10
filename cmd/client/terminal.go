package main

import (
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorGrey   = "\033[38;5;243m"
	colorOrange = "\033[38;5;214m"
	colorWhite  = "\033[38;5;255m"
	colorBold   = "\033[1m"
)

// promptPassword displays a styled password prompt and reads the password with visual feedback
// Platform-specific implementation in terminal_unix.go, terminal_freebsd.go, and terminal_windows.go
func promptPassword(message string, optional bool) (string, error) {
	// Build the prompt with styling
	var prompt strings.Builder
	prompt.WriteString(colorOrange)
	prompt.WriteString(colorBold)
	prompt.WriteString("→ ")
	prompt.WriteString(colorReset)
	prompt.WriteString(colorWhite)
	prompt.WriteString(message)
	if optional {
		prompt.WriteString(" ")
		prompt.WriteString(colorGrey)
		prompt.WriteString("(or press Enter for none)")
		prompt.WriteString(colorReset)
	}
	prompt.WriteString(colorWhite)
	prompt.WriteString(": ")
	prompt.WriteString(colorReset)

	fmt.Print(prompt.String())

	password, err := readPasswordWithBullets(prompt.String())
	if err != nil {
		return "", err
	}

	fmt.Println() // New line after password entry
	return password, nil
}
