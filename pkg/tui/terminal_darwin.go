//go:build darwin

package tui

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	// FreeBSD terminal ioctls
	TIOCGETA = 0x402c7413
	TIOCSETA = 0x802c7414
)

// readPasswordWithBullets reads a password with visual feedback (bullets)
func readPasswordWithBullets(prompt string, showBullets bool) (string, error) {
	var oldState syscall.Termios
	fd := int(os.Stdin.Fd())

	// Get terminal state using FreeBSD TIOCGETA
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), TIOCGETA, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
		return "", fmt.Errorf("failed to get terminal state: %v", err)
	}

	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON
	newState.Lflag |= syscall.ISIG
	newState.Iflag &^= syscall.ICRNL

	// Set terminal state using FreeBSD TIOCSETA
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), TIOCSETA, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return "", fmt.Errorf("failed to set terminal state: %v", err)
	}

	defer func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), TIOCSETA, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
	}()

	var buf [1]byte
	var password []byte

	for {
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			return "", err
		}

		if n == 0 || buf[0] == '\n' || buf[0] == '\r' {
			break
		}

		if buf[0] == 127 || buf[0] == 8 { // Backspace
			if len(password) > 0 {
				password = password[:len(password)-1]
				if showBullets && prompt != "" {
					// Clear the line and reprint
					fmt.Print("\r")
					fmt.Print(prompt)
					fmt.Print(termColorGrey)
					fmt.Print(strings.Repeat("•", len(password)))
					fmt.Print(termColorReset)
				}
			}
			continue
		}

		password = append(password, buf[0])
		if showBullets {
			// Print grey bullet
			fmt.Print(termColorGrey)
			fmt.Print("•")
			fmt.Print(termColorReset)
		}
	}

	return string(password), nil
}
