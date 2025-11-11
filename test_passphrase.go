package main

import (
	"fmt"
	"github.com/r2unit/openpasswd/pkg/config"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"os"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	passwords, _ := db.ListPasswords()
	fmt.Printf("Database has %d passwords\n", len(passwords))

	if len(passwords) > 0 {
		fmt.Printf("\nFirst password details:\n")
		p := passwords[0]
		fmt.Printf("  ID: %d\n", p.ID)
		fmt.Printf("  Name (encrypted): %q\n", p.Name[:20])
		fmt.Printf("  Name length: %d\n", len(p.Name))
	}

	// Try to decrypt with test passphrase
	testPass := "test123"
	encryptor := crypto.NewEncryptorWithVersion(testPass, cfg.Salt, cfg.KDFVersion)
	
	fmt.Printf("\nTrying to decrypt with passphrase: %q\n", testPass)
	
	if len(passwords) > 0 {
		decrypted, err := encryptor.Decrypt(passwords[0].Name)
		if err != nil {
			fmt.Printf("  Decryption FAILED: %v\n", err)
		} else {
			fmt.Printf("  Decryption SUCCESS: %q\n", decrypted)
		}
	}
}
