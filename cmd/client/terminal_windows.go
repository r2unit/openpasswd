//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	enableLineInput = 0x0002
	enableEchoInput = 0x0004
)

// readPasswordWithBullets reads a password with visual feedback (bullets)
func readPasswordWithBullets(prompt string) (string, error) {
	// On Windows, we'll use a simpler approach with bufio
	// since low-level console manipulation is complex and requires syscalls
	// For now, fall back to reading line without bullets on Windows

	reader := bufio.NewReader(os.Stdin)

	// Try to disable echo on Windows console
	handle := syscall.Handle(os.Stdin.Fd())
	var mode uint32

	r, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r != 0 {
		// Success - disable echo
		newMode := mode &^ (enableEchoInput | enableLineInput)
		procSetConsoleMode.Call(uintptr(handle), uintptr(newMode))

		defer func() {
			// Restore original mode
			procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
		}()

		var password []byte
		var buf [1]byte

		for {
			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				return "", err
			}

			if n == 0 || buf[0] == '\n' || buf[0] == '\r' {
				break
			}

			if buf[0] == 8 { // Backspace on Windows
				if len(password) > 0 {
					password = password[:len(password)-1]
					fmt.Print("\b \b") // Backspace, space, backspace
				}
				continue
			}

			password = append(password, buf[0])
			fmt.Print(colorGrey + "•" + colorReset)
		}

		return string(password), nil
	}

	// Fallback: if console mode manipulation fails, just read the line
	// This happens in non-console environments (pipes, IDE terminals, etc.)
	if err != nil {
		// Silently use fallback
		_ = err
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(line, "\r\n"), nil
}
