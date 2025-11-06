package main

import (
	"fmt"
	"os"

	"github.com/r2unit/openpasswd/pkg/config"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/mfa"
	_ "github.com/r2unit/openpasswd/pkg/proton/pass" // Register Proton Pass provider
	"github.com/r2unit/openpasswd/pkg/tui"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		initializeConfig()
	case "server":
		fmt.Fprintf(os.Stderr, tui.ColorWarning("Server mode is currently disabled\n"))
		os.Exit(1)
	case "auth":
		handleAuth()
	case "add":
		handleAdd()
	case "list":
		handleList()
	case "settings":
		handleSettings()
	case "import":
		handleImport()
	case "help", "--help", "-h":
		showHelp()
	default:
		showHelp()
	}
}

func initializeConfig() {
	fmt.Println(tui.ColorInfo("Initializing OpenPasswd..."))

	salt, err := crypto.GenerateSalt()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error generating salt: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveSalt(salt); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving configuration: %v\n", err)))
		os.Exit(1)
	}

	if err := config.CreateDefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorWarning(fmt.Sprintf("Warning: Could not create config.toml: %v\n", err)))
	}

	configDir, _ := config.GetConfigDir()
	fmt.Println(tui.ColorSuccess("Configuration initialized successfully!"))
	fmt.Printf("Config directory: %s\n", configDir)
	fmt.Printf("Config file: %s/config.toml\n", configDir)
	fmt.Printf("\n%s\n", tui.ColorInfo("Run 'openpass' to start the password manager"))
	fmt.Printf("%s\n", tui.ColorInfo("Edit config.toml to customize colors"))
}

func showHelp() {
	help := `OpenPasswd - A secure, terminal-based password manager

COMMANDS:
    openpasswd init              Initialize configuration and database
    openpasswd add               Add a new password entry
    openpasswd list              List and search passwords
    openpasswd auth              Connect to password providers (Proton Pass, etc.)
    openpasswd import            Import passwords from other password managers
    openpasswd settings          Manage settings (passphrase, MFA, etc.)
    openpasswd help              Show this help message

OPTIONS:
    --help, -h                   Show this help message

EXAMPLES:
    openpasswd init                             # First-time setup
    openpasswd add                              # Add password interactively
    openpasswd add login                        # Add login password
    openpasswd list                             # List all passwords
    openpasswd auth login                       # Connect to Proton Pass and sync passwords
    openpasswd import                           # Import from export files
    openpasswd settings set-passphrase          # Set master passphrase
    openpasswd settings set-totp                # Enable TOTP authentication
    openpasswd settings set-yubikey             # Enable YubiKey authentication

CONFIGURATION:
    ~/.config/openpasswd/passwords.db    Encrypted password database
    ~/.config/openpasswd/salt            Encryption salt
    ~/.config/openpasswd/passphrase      Master passphrase (optional)
    ~/.config/openpasswd/totp_secret     TOTP secret (optional)
    ~/.config/openpasswd/config.toml     Color configuration

For more information, visit: https://github.com/r2unit/openpasswd
`
	fmt.Println(help)
}

func handleImport() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nRun 'openpass init' to initialize the password manager")
		os.Exit(1)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	passphrase := ""
	if config.HasPassphrase() {
		fmt.Print("Enter master passphrase: ")
		passphrase, err = readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println()

		savedPass, err := config.LoadPassphrase()
		if err != nil || savedPass != passphrase {
			fmt.Fprintf(os.Stderr, tui.ColorError("Incorrect passphrase\n"))
			os.Exit(1)
		}
	}

	if err := tui.RunImportTUI(db, cfg.Salt, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

func handleAuth() {
	if len(os.Args) < 3 {
		showAuthHelp()
		return
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "login":
		handleAuthLogin()
	case "logout":
		handleAuthLogout()
	case "status":
		handleAuthStatus()
	case "help", "--help", "-h":
		showAuthHelp()
	default:
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Unknown auth command: %s\n", subcommand)))
		showAuthHelp()
		os.Exit(1)
	}
}

func handleAuthLogin() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nRun 'openpasswd init' to initialize the password manager")
		os.Exit(1)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	passphrase := ""
	if config.HasPassphrase() {
		fmt.Print("Enter master passphrase: ")
		passphrase, err = readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println()

		savedPass, err := config.LoadPassphrase()
		if err != nil || savedPass != passphrase {
			fmt.Fprintf(os.Stderr, tui.ColorError("Incorrect passphrase\n"))
			os.Exit(1)
		}
	}

	if err := tui.RunAuthLoginTUI(db, cfg.Salt, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

func handleAuthLogout() {
	fmt.Println(tui.ColorInfo("Auth logout functionality coming soon"))
	fmt.Println(tui.ColorInfo("For now, passwords are synced locally only"))
}

func handleAuthStatus() {
	fmt.Println(tui.ColorInfo("Auth status functionality coming soon"))
	fmt.Println(tui.ColorInfo("For now, use 'openpasswd list' to see your synced passwords"))
}

func showAuthHelp() {
	help := `OpenPasswd - Auth Command

COMMANDS:
    openpasswd auth login         Connect to a password provider and sync passwords
    openpasswd auth logout        Disconnect from provider (coming soon)
    openpasswd auth status        Show current auth status (coming soon)
    openpasswd auth help          Show this help message

DESCRIPTION:
    The auth command allows you to connect to external password providers
    and sync your passwords to openpasswd. Currently supported providers:
    
    - Proton Pass (via export file)
    
    More providers coming soon:
    - Bitwarden
    - 1Password
    - LastPass

EXAMPLES:
    openpasswd auth login         # Show list of available providers
    openpasswd auth status        # Check connection status
    openpasswd auth logout        # Disconnect from provider

WORKFLOW:
    1. Export your passwords from your current password manager
    2. Run 'openpasswd auth login'
    3. Select your provider from the list
    4. Enter required credentials (usually export file path)
    5. Your passwords will be synced and encrypted locally
`
	fmt.Println(help)
}

func handleAdd() {
	if len(os.Args) >= 3 && (os.Args[2] == "help" || os.Args[2] == "--help" || os.Args[2] == "-h") {
		showAddHelp()
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nRun 'openpass init' to initialize the password manager")
		os.Exit(1)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	passwordType := ""
	if len(os.Args) >= 3 {
		passwordType = os.Args[2]
	}

	passphrase := ""

	if err := tui.RunAddTUI(db, cfg.Salt, passphrase, passwordType); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

func showAddHelp() {
	help := `OpenPasswd - Add Password Command

COMMANDS:
    openpass add                               Add password interactively
    openpass add <name>                        Add password with name
    openpass add <name> <username>             Add password with name and username
    openpass add <name> <username> <url>       Add password with name, username, and URL
    openpass add help                          Show this help message

OPTIONS:
    --help, -h                                 Show this help message

DESCRIPTION:
    Add a new password entry to the password manager. You can provide
    arguments on the command line or enter them interactively.
    
    The password will always be prompted securely (hidden input).
    URL and notes are optional fields.

EXAMPLES:
    openpass add                               # Interactive mode
    openpass add "GitHub"                      # Add with name only
    openpass add "GitHub" "myuser"             # Add with name and username
    openpass add "GitHub" "myuser" "github.com"  # Add with name, user, and URL
    openpass add --help                        # Show this help

REQUIRED:
    Name         A descriptive name for the password entry
    Username     The username or email for the account
    Password     The password (prompted securely)

OPTIONAL:
    URL          The website or service URL
    Notes        Additional notes or information
`
	fmt.Println(help)
}

func readPassword() (string, error) {
	password, err := readPasswordFromTerminal()
	if err != nil {
		return "", err
	}
	return password, nil
}

func handleSettings() {
	if len(os.Args) >= 3 && (os.Args[2] == "help" || os.Args[2] == "--help" || os.Args[2] == "-h") {
		showSettingsHelp()
		return
	}

	_, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nRun 'openpass init' to initialize the password manager")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		showSettingsHelp()
		return
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "set-passphrase":
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error loading config: %v\n", err)))
			os.Exit(1)
		}

		db, err := database.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error opening database: %v\n", err)))
			os.Exit(1)
		}
		defer db.Close()

		passwords, _ := db.ListPasswords()
		oldPassphrase := ""

		if config.HasPassphrase() {
			fmt.Print("Enter current master passphrase: ")
			oldPassphrase, err = readPassword()
			if err != nil {
				fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
				os.Exit(1)
			}
			fmt.Println()

			oldPass, err := config.LoadPassphrase()
			if err != nil || oldPass != oldPassphrase {
				fmt.Fprintf(os.Stderr, tui.ColorError("Incorrect passphrase\n"))
				os.Exit(1)
			}
		}

		fmt.Print("Enter new master passphrase: ")
		newPassphrase, err := readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println()

		fmt.Print("Confirm new master passphrase: ")
		confirm, err := readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println()

		if newPassphrase != confirm {
			fmt.Fprintf(os.Stderr, tui.ColorError("Passphrases do not match\n"))
			os.Exit(1)
		}

		if len(passwords) > 0 {
			fmt.Println(tui.ColorInfo(fmt.Sprintf("Re-encrypting %d passwords with new passphrase...", len(passwords))))

			oldEncryptor := crypto.NewEncryptor(oldPassphrase, cfg.Salt)
			newEncryptor := crypto.NewEncryptor(newPassphrase, cfg.Salt)

			for _, p := range passwords {
				if p.Name != "" {
					decrypted, err := oldEncryptor.Decrypt(p.Name)
					if err != nil {
						fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error decrypting password ID %d: %v\n", p.ID, err)))
						os.Exit(1)
					}
					p.Name, _ = newEncryptor.Encrypt(decrypted)
				}

				if p.Username != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.Username); err == nil {
						p.Username, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if p.Password != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.Password); err == nil {
						p.Password, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if p.URL != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.URL); err == nil {
						p.URL, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if p.Notes != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.Notes); err == nil {
						p.Notes, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				for key, val := range p.Fields {
					if decrypted, err := oldEncryptor.Decrypt(val); err == nil {
						p.Fields[key], _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if err := db.UpdatePassword(p); err != nil {
					fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error updating password ID %d: %v\n", p.ID, err)))
					os.Exit(1)
				}
			}
		}

		if err := config.SavePassphrase(newPassphrase); err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving passphrase: %v\n", err)))
			os.Exit(1)
		}

		fmt.Println(tui.ColorSuccess("Master passphrase set successfully!"))
		if len(passwords) > 0 {
			fmt.Println(tui.ColorSuccess(fmt.Sprintf("Re-encrypted %d passwords", len(passwords))))
		}

	case "remove-passphrase":
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error loading config: %v\n", err)))
			os.Exit(1)
		}

		if !config.HasPassphrase() {
			fmt.Println(tui.ColorWarning("No passphrase is currently set"))
			return
		}

		fmt.Print("Enter current master passphrase: ")
		oldPassphrase, err := readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println()

		savedPass, err := config.LoadPassphrase()
		if err != nil || savedPass != oldPassphrase {
			fmt.Fprintf(os.Stderr, tui.ColorError("Incorrect passphrase\n"))
			os.Exit(1)
		}

		db, err := database.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error opening database: %v\n", err)))
			os.Exit(1)
		}
		defer db.Close()

		passwords, _ := db.ListPasswords()

		if len(passwords) > 0 {
			fmt.Println(tui.ColorWarning(fmt.Sprintf("Warning: This will re-encrypt %d passwords without passphrase protection", len(passwords))))
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Operation cancelled")
				return
			}

			fmt.Println(tui.ColorInfo(fmt.Sprintf("Re-encrypting %d passwords...", len(passwords))))

			oldEncryptor := crypto.NewEncryptor(oldPassphrase, cfg.Salt)
			newEncryptor := crypto.NewEncryptor("", cfg.Salt)

			for _, p := range passwords {
				if p.Name != "" {
					decrypted, err := oldEncryptor.Decrypt(p.Name)
					if err != nil {
						fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error decrypting password ID %d: %v\n", p.ID, err)))
						os.Exit(1)
					}
					p.Name, _ = newEncryptor.Encrypt(decrypted)
				}

				if p.Username != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.Username); err == nil {
						p.Username, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if p.Password != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.Password); err == nil {
						p.Password, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if p.URL != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.URL); err == nil {
						p.URL, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if p.Notes != "" {
					if decrypted, err := oldEncryptor.Decrypt(p.Notes); err == nil {
						p.Notes, _ = newEncryptor.Encrypt(decrypted)
					}
				}

				for key, val := range p.Fields {
					if decrypted, err := oldEncryptor.Decrypt(val); err == nil {
						p.Fields[key], _ = newEncryptor.Encrypt(decrypted)
					}
				}

				if err := db.UpdatePassword(p); err != nil {
					fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error updating password ID %d: %v\n", p.ID, err)))
					os.Exit(1)
				}
			}

			fmt.Println(tui.ColorSuccess(fmt.Sprintf("Re-encrypted %d passwords", len(passwords))))
		}

		if err := config.RemovePassphrase(); err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error removing passphrase: %v\n", err)))
			os.Exit(1)
		}

		fmt.Println(tui.ColorSuccess("Master passphrase removed successfully!"))
		fmt.Println(tui.ColorWarning("Note: Passwords no longer require passphrase to access"))

	case "set-totp":
		handleSetTOTP()

	case "remove-totp":
		handleRemoveTOTP()

	case "show-totp-qr":
		handleShowTOTPQR()

	case "set-yubikey":
		handleSetYubiKey()

	case "remove-yubikey":
		handleRemoveYubiKey()

	default:
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Unknown settings command: %s\n", subcommand)))
		showSettingsHelp()
		os.Exit(1)
	}
}

func handleList() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("\nRun 'openpass init' to initialize the password manager")
		os.Exit(1)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	passphrase := ""
	if config.HasPassphrase() {
		fmt.Print("Enter master passphrase: ")
		passphrase, err = readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println()

		savedPass, err := config.LoadPassphrase()
		if err != nil || savedPass != passphrase {
			fmt.Fprintf(os.Stderr, tui.ColorError("Incorrect passphrase\n"))
			os.Exit(1)
		}
	}

	if err := tui.RunListTUI(db, cfg.Salt, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

func handleSetTOTP() {
	username := "user"

	if err := tui.RunTOTPSetupTUI(username); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("\n✓ TOTP authentication enabled successfully!"))
	fmt.Println(tui.ColorInfo("You will need to provide a TOTP code when accessing passwords."))
}

func handleRemoveTOTP() {
	if !config.HasTOTP() {
		fmt.Println(tui.ColorWarning("TOTP authentication is not currently enabled"))
		return
	}

	fmt.Print("Are you sure you want to remove TOTP authentication? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("Operation cancelled")
		return
	}

	if err := config.RemoveTOTP(); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error removing TOTP: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("✓ TOTP authentication removed successfully!"))
}

func handleShowTOTPQR() {
	if !config.HasTOTP() {
		fmt.Println(tui.ColorWarning("TOTP authentication is not currently enabled"))
		fmt.Println("Run 'openpass settings set-totp' to enable it")
		return
	}

	fmt.Println(tui.ColorWarning("Re-displaying QR code is not supported"))
	fmt.Println(tui.ColorInfo("Please remove and re-add TOTP to get a new QR code:"))
	fmt.Println("  openpass settings remove-totp")
	fmt.Println("  openpass settings set-totp")
}

func handleSetYubiKey() {
	if !mfa.IsYubiKeyAvailable() {
		fmt.Fprintf(os.Stderr, tui.ColorError("YubiKey not detected or ykman not installed\n"))
		fmt.Println(tui.ColorInfo("\nTo use YubiKey:"))
		fmt.Println("1. Install ykman from https://www.yubico.com/")
		fmt.Println("2. Insert your YubiKey")
		fmt.Println("3. Run this command again")
		os.Exit(1)
	}

	if err := mfa.ConfigureYubiKey(); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error configuring YubiKey: %v\n", err)))
		os.Exit(1)
	}

	challenge, err := mfa.GenerateYubiKeyChallenge()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error generating challenge: %v\n", err)))
		os.Exit(1)
	}

	if err := mfa.TestYubiKey(challenge); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("YubiKey test failed: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveYubiKeyChallenge(challenge); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving YubiKey config: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("✓ YubiKey authentication enabled successfully!"))
	fmt.Println(tui.ColorInfo("You will need your YubiKey present when accessing passwords."))
}

func handleRemoveYubiKey() {
	if !config.HasYubiKey() {
		fmt.Println(tui.ColorWarning("YubiKey authentication is not currently enabled"))
		return
	}

	fmt.Print("Are you sure you want to remove YubiKey authentication? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Println("Operation cancelled")
		return
	}

	if err := config.RemoveYubiKey(); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error removing YubiKey: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("✓ YubiKey authentication removed successfully!"))
}

func showSettingsHelp() {
	help := `OpenPasswd - Settings Command

COMMANDS:
    openpass settings set-passphrase      Set a master passphrase (optional)
    openpass settings remove-passphrase   Remove the master passphrase
    openpass settings set-totp            Enable TOTP (authenticator app)
    openpass settings remove-totp         Disable TOTP authentication
    openpass settings show-totp-qr        Show TOTP QR code again
    openpass settings set-yubikey         Enable YubiKey authentication
    openpass settings remove-yubikey      Disable YubiKey authentication
    openpass settings help                Show this help message

DESCRIPTION:
    Configure OpenPasswd authentication methods. You can combine multiple
    methods for stronger security:
    - Passphrase: Simple password protection
    - TOTP: Time-based codes from authenticator app
    - YubiKey: Hardware key authentication

EXAMPLES:
    openpass settings set-passphrase      # Set a master passphrase
    openpass settings set-totp            # Enable Google Authenticator
    openpass settings set-yubikey         # Enable YubiKey
    openpass settings show-totp-qr        # Re-display QR code
`
	fmt.Println(help)
}
