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
	
	if len(passwords) == 0 {
		fmt.Println("No passwords in database")
		return
	}

	// Try with empty passphrase (the bug we suspected)
	fmt.Println("Testing with EMPTY passphrase:")
	emptyEncryptor := crypto.NewEncryptorWithVersion("", cfg.Salt, cfg.KDFVersion)
	decrypted, err := emptyEncryptor.Decrypt(passwords[0].Name)
	if err != nil {
		fmt.Printf("  Empty passphrase FAILED: %v\n", err)
	} else {
		fmt.Printf("  Empty passphrase SUCCESS! Name: %q\n", decrypted)
	}
}
