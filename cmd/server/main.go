package main

import (
	"fmt"
	"os"

	"github.com/r2unit/openpasswd/pkg/version"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "start":
		handleStart()
	case "version", "--version", "-v":
		handleVersion()
	case "help", "--help", "-h":
		showHelp()
	default:
		showHelp()
	}
}

func showHelp() {
	help := `OpenPasswd Server - Sync server for OpenPasswd password manager

COMMANDS:
    openpasswd-server start              Start the sync server
    openpasswd-server version            Show version information
    openpasswd-server help               Show this help message

OPTIONS:
    --help, -h                           Show this help message
    --version, -v                        Show version number

DESCRIPTION:
    The OpenPasswd server provides sync capabilities for the OpenPasswd
    password manager, allowing you to sync your passwords across multiple
    devices securely.

EXAMPLES:
    openpasswd-server start              # Start the sync server
    openpasswd-server version            # Show version info

CONFIGURATION:
    Server configuration will use the same config directory as the client:
    ~/.config/openpasswd/

FUTURE FEATURES:
    - End-to-end encrypted sync
    - Multi-device support
    - Conflict resolution
    - RESTful API
    - WebSocket support for real-time sync

For more information, visit: https://github.com/r2unit/openpasswd
`
	fmt.Println(help)
}

func handleStart() {
	fmt.Println("OpenPasswd Server - Starting...")
	fmt.Println()
	fmt.Println("⚠ Server functionality is currently under development")
	fmt.Println()
	fmt.Println("The sync server will provide:")
	fmt.Println("  • End-to-end encrypted synchronization")
	fmt.Println("  • Multi-device support")
	fmt.Println("  • Conflict resolution")
	fmt.Println("  • RESTful API for client connections")
	fmt.Println()
	fmt.Println("Coming soon in a future release!")
}

func handleVersion() {
	info := version.GetInfo()

	// Display basic version
	fmt.Printf("OpenPasswd Server v%s\n", info.Version)

	// Check for --verbose flag
	if len(os.Args) > 2 && (os.Args[2] == "--verbose" || os.Args[2] == "-v") {
		fmt.Printf("\nBuild Information:\n")
		fmt.Printf("  Git Commit:  %s\n", info.GitCommit)
		fmt.Printf("  Build Date:  %s\n", info.BuildDate)
		fmt.Printf("  Go Version:  %s\n", info.GoVersion)
		fmt.Printf("  Platform:    %s\n", info.Platform)
	}
}
