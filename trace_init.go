package main

import (
	"fmt"
	"os"
	"github.com/r2unit/openpasswd/pkg/tui"
)

func main() {
	fmt.Println("Running setup TUI...")
	result, err := tui.RunSetupTUI()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	if result.Cancelled {
		fmt.Println("Setup cancelled")
		os.Exit(0)
	}
	
	fmt.Printf("\nSetup complete!\n")
	fmt.Printf("Passphrase length: %d\n", len(result.Passphrase))
	fmt.Printf("Passphrase bytes: %v\n", []byte(result.Passphrase))
	fmt.Printf("Passphrase repr: %q\n", result.Passphrase)
}
