//go:build linux

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// readPasswordWithBullets reads a password with visual feedback (bullets)
func readPasswordWithBullets(prompt string) (string, error) {
	var oldState syscall.Termios
	fd := int(os.Stdin.Fd())

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
		return "", fmt.Errorf("failed to get terminal state: %v", err)
	}

	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON
	newState.Lflag |= syscall.ISIG
	newState.Iflag &^= syscall.ICRNL

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return "", fmt.Errorf("failed to set terminal state: %v", err)
	}

	defer func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
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
				// Clear the line and reprint
				fmt.Print("\r")
				fmt.Print(prompt)
				fmt.Print(colorGrey)
				fmt.Print(strings.Repeat("•", len(password)))
				fmt.Print(colorReset)
			}
			continue
		}

		password = append(password, buf[0])
		// Print grey bullet
		fmt.Print(colorGrey)
		fmt.Print("•")
		fmt.Print(colorReset)
	}

	return string(password), nil
}
