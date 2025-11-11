package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/r2unit/openpasswd/pkg/config"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/mfa"
	_ "github.com/r2unit/openpasswd/pkg/proton/pass" // Register Proton Pass provider
	"github.com/r2unit/openpasswd/pkg/tui"
	"github.com/r2unit/openpasswd/pkg/version"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	cmd := os.Args[1]

	// Commands that don't require initialization
	switch cmd {
	case "init":
		initializeConfig()
		return
	case "version", "--version", "-v":
		handleVersion()
		return
	case "upgrade":
		handleUpgrade()
		return
	case "help", "--help", "-h":
		showHelp()
		return
	}

	// Check for updates on startup (non-blocking, cached)
	checkForUpdatesStartup()

	// All other commands require initialization
	if !isInitialized() {
		showNotInitializedError()
		os.Exit(1)
	}

	// Commands that require initialization
	switch cmd {
	case "server":
		fmt.Fprintf(os.Stderr, "%s", tui.ColorWarning("Server mode is currently disabled\n"))
		os.Exit(1)

	case "auth":
		fmt.Fprintf(os.Stderr, "%s", tui.ColorWarning("Auth functionality is currently disabled\n"))
		os.Exit(1)

	case "import":
		fmt.Fprintf(os.Stderr, "%s", tui.ColorWarning("Import functionality is currently disabled\n"))
		os.Exit(1)

	case "add":
		handleAdd()
	case "list":
		handleList()
	case "settings":
		handleSettings()
	case "migrate":
		handleMigrate()
	default:
		showHelp()
	}
}

// isInitialized checks if OpenPasswd has been initialized
func isInitialized() bool {
	_, err := config.LoadConfig()
	return err == nil
}

// showNotInitializedError displays a helpful error message
func showNotInitializedError() {
	fmt.Println()
	fmt.Println(tui.ColorError("✗ OpenPasswd has not been initialized yet!"))
	fmt.Println()
	fmt.Println(tui.ColorInfo("Please run the following command to set up OpenPasswd:"))
	fmt.Println()
	fmt.Println(tui.ColorSuccess("  openpasswd init"))
	fmt.Println()
	fmt.Println(tui.ColorInfo("This will:"))
	fmt.Println(tui.ColorInfo("  • Create your master passphrase"))
	fmt.Println(tui.ColorInfo("  • Generate a recovery key"))
	fmt.Println(tui.ColorInfo("  • Set up secure encryption"))
	fmt.Println()
}

func initializeConfig() {
	// Check if configuration already exists
	if _, err := config.LoadConfig(); err == nil {
		// Configuration exists, prompt user with TUI (fallback to simple prompts if TTY not available)
		choice, err := tui.RunInitTUI()

		// If TUI fails (no TTY), use simple text prompts as fallback
		if err != nil {
			choice = initPromptFallback()
		}

		switch choice {
		case tui.ChoiceIgnore:
			fmt.Println(tui.ColorInfo("\n✓ Keeping existing configuration."))
			configDir, _ := config.GetConfigDir()
			fmt.Printf("  Config directory: %s\n", configDir)
			return
		case tui.ChoiceOverride:
			fmt.Println(tui.ColorInfo("\n⚠ Overriding existing configuration..."))
		case tui.ChoiceCancel:
			fmt.Println(tui.ColorInfo("\nOperation cancelled."))
			os.Exit(0)
		}
	}

	// Run setup TUI to get passphrase and recovery key
	setupResult, err := tui.RunSetupTUI()
	if err != nil || setupResult.Cancelled {
		fmt.Println(tui.ColorInfo("\nSetup cancelled."))
		os.Exit(0)
	}

	// Generate salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error generating salt: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveSalt(salt); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving configuration: %v\n", err)))
		os.Exit(1)
	}

	// Save current KDF version (600k iterations)
	if err := config.SaveKDFVersion(crypto.CurrentKDFVersion); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving KDF version: %v\n", err)))
		os.Exit(1)
	}

	// Encrypt and save recovery key
	encryptedRecovery, err := crypto.EncryptRecoveryKey(setupResult.RecoveryKey, setupResult.Passphrase, salt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error encrypting recovery key: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveRecoveryKey(encryptedRecovery); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving recovery key: %v\n", err)))
		os.Exit(1)
	}

	// Save recovery key hash for verification
	recoveryHash := crypto.GenerateRecoveryHash(setupResult.RecoveryKey)
	if err := config.SaveRecoveryHash(crypto.EncodeRecoveryHash(recoveryHash)); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving recovery hash: %v\n", err)))
		os.Exit(1)
	}

	if err := config.CreateDefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorWarning(fmt.Sprintf("Warning: Could not create config.toml: %v\n", err)))
	}

	configDir, _ := config.GetConfigDir()
	fmt.Println(tui.ColorSuccess("\n✓ Configuration initialized successfully!"))
	fmt.Printf("  Config directory: %s\n", configDir)
	fmt.Printf("  Config file: %s/config.toml\n", configDir)
	fmt.Println()
	fmt.Println(tui.ColorWarning("⚠  Your recovery key has been saved encrypted."))
	fmt.Println(tui.ColorInfo("  Keep your handwritten/backup copy in a safe place!"))
	fmt.Printf("\n%s\n", tui.ColorInfo("Run 'openpasswd list' to start using the password manager"))
}

func initPromptFallback() tui.InitChoice {
	fmt.Println(tui.ColorWarning("\n⚠ Configuration already exists!"))
	fmt.Println("\nWhat would you like to do?")
	fmt.Println("  1) Ignore (keep existing configuration)")
	fmt.Println("  2) Override (replace with new configuration)")
	fmt.Print("\nEnter choice (1 or 2): ")

	var choice string
	fmt.Scanln(&choice)

	if choice == "1" {
		return tui.ChoiceIgnore
	} else if choice == "2" {
		// Extra confirmation for override
		fmt.Println(tui.ColorWarning("\n⚠ WARNING: This will delete your existing configuration and database!"))
		fmt.Println(tui.ColorWarning("All stored passwords will be lost."))
		fmt.Print("\nType 'yes' to confirm override: ")

		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" {
			return tui.ChoiceCancel
		}

		return tui.ChoiceOverride
	}

	return tui.ChoiceCancel
}

func showHelp() {
	help := `OpenPasswd - A secure, terminal-based password manager

COMMANDS:
    openpasswd init              Initialize configuration and database
    openpasswd add               Add a new password entry
    openpasswd list              List and search passwords
    openpasswd settings          Manage settings (passphrase, MFA, etc.)
    openpasswd version           Show version information
    openpasswd upgrade           Upgrade to the latest version
    openpasswd help              Show this help message

OPTIONS:
    --help, -h                   Show this help message
    --version, -v                Show version number

EXAMPLES:
    openpasswd init                             # First-time setup
    openpasswd add                              # Add password interactively
    openpasswd add login                        # Add login password
    openpasswd list                             # List all passwords
    openpasswd settings set-passphrase          # Set master passphrase
    openpasswd settings set-totp                # Enable TOTP authentication
    openpasswd settings set-yubikey             # Enable YubiKey authentication
    openpasswd version --verbose                # Show detailed version info
    openpasswd version --check                  # Check for updates (interactive)
    openpasswd version --disable-checking       # Disable automatic update checks
    openpasswd version --enable-checking        # Enable automatic update checks
    openpasswd upgrade                          # Upgrade to latest version

CONFIGURATION:
    ~/.config/openpasswd/passwords.db          Encrypted password database
    ~/.config/openpasswd/salt                  Encryption salt
    ~/.config/openpasswd/totp_secret           TOTP secret (optional)
    ~/.config/openpasswd/config.toml           Color configuration
    ~/.config/openpasswd/disable_version_check Flag to disable auto-update checks
    ~/.cache/openpasswd/version_check.json     Cached version check (24hr TTL)

For more information, visit: https://github.com/r2unit/openpasswd
`
	fmt.Println(help)
}

// This function will allow importing passwords from export files of various password managers
// Planned support: Bitwarden, 1Password, LastPass, KeePass, etc.
// Currently disabled - implementation pending

// This function will handle OAuth/API authentication with providers like:
// - Proton Pass (via API)
// - Bitwarden (self-hosted or cloud)
// - 1Password (via CLI integration)
// Currently disabled - implementation pending

// Will implement OAuth2 flow or API key authentication
// Should securely store access tokens using system keychain

// Will revoke access tokens and clear stored credentials

// Will display connected providers and sync status

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

	// Always prompt for passphrase
	passphrase, err := promptPassword("Enter master passphrase", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error reading passphrase: %v\n", err)))
		os.Exit(1)
	}

	// Validate passphrase
	if !validatePassphrase(db, cfg.Salt, passphrase, cfg.KDFVersion) {
		if err := tui.RunWrongPassphraseTUI(); err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError("Error: %v\n"), err)
		}
		os.Exit(1)
	}

	if err := tui.RunAddTUI(db, cfg.Salt, passphrase, passwordType); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
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
		fmt.Println(tui.ColorWarning("⚠ This feature has been removed for security reasons."))
		fmt.Println(tui.ColorInfo("Storing your master passphrase on disk defeats the purpose of encryption."))
		fmt.Println(tui.ColorInfo("You will be prompted for your passphrase when needed."))
		fmt.Println()
		fmt.Println(tui.ColorInfo("Your passwords are still encrypted and secure!"))
		fmt.Println(tui.ColorInfo("Simply enter your passphrase when running commands like 'openpasswd list'."))

	case "remove-passphrase":
		fmt.Println(tui.ColorWarning("⚠ This feature has been removed for security reasons."))
		fmt.Println(tui.ColorInfo("Passphrase storage on disk is no longer supported."))
		fmt.Println()
		fmt.Println(tui.ColorInfo("If you previously stored a passphrase, you can delete it manually:"))
		configDir, _ := config.GetConfigDir()
		fmt.Printf(tui.ColorInfo("  rm %s/passphrase\n"), configDir)

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
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Unknown settings command: %s\n", subcommand)))
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

	// Always prompt for passphrase (plaintext storage removed for security)
	passphrase, err := promptPassword("Enter master passphrase", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error reading passphrase: %v\n", err)))
		os.Exit(1)
	}

	// Validate passphrase before showing TUI
	if !validatePassphrase(db, cfg.Salt, passphrase, cfg.KDFVersion) {
		if err := tui.RunWrongPassphraseTUI(); err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError("Error: %v\n"), err)
		}
		os.Exit(1)
	}

	if err := tui.RunListTUI(db, cfg.Salt, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

// validatePassphrase checks if the passphrase can decrypt the database
func validatePassphrase(db *database.DB, salt []byte, passphrase string, kdfVersion int) bool {
	passwords, err := db.ListPasswords()
	if err != nil {
		return false
	}

	// If database is empty, any passphrase is "valid" (no way to verify)
	if len(passwords) == 0 {
		return true
	}

	// Try to decrypt the first password's name using the correct KDF version
	encryptor := crypto.NewEncryptorWithVersion(passphrase, salt, kdfVersion)
	for _, p := range passwords {
		if p.Name != "" {
			_, err := encryptor.Decrypt(p.Name)
			// If decryption succeeds, passphrase is correct
			return err == nil
		}
	}

	// If no encrypted fields found, assume passphrase is valid
	return true
}

func handleSetTOTP() {
	username := "user"

	if err := tui.RunTOTPSetupTUI(username); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
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
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error removing TOTP: %v\n", err)))
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
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError("YubiKey not detected or ykman not installed\n"))
		fmt.Println(tui.ColorInfo("\nTo use YubiKey:"))
		fmt.Println("1. Install ykman from https://www.yubico.com/")
		fmt.Println("2. Insert your YubiKey")
		fmt.Println("3. Run this command again")
		os.Exit(1)
	}

	if err := mfa.ConfigureYubiKey(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error configuring YubiKey: %v\n", err)))
		os.Exit(1)
	}

	challenge, err := mfa.GenerateYubiKeyChallenge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error generating challenge: %v\n", err)))
		os.Exit(1)
	}

	if err := mfa.TestYubiKey(challenge); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("YubiKey test failed: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveYubiKeyChallenge(challenge); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving YubiKey config: %v\n", err)))
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
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error removing YubiKey: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("✓ YubiKey authentication removed successfully!"))
}

func showSettingsHelp() {
	help := `OpenPasswd - Settings Command

COMMANDS:
    openpass settings set-totp            Enable TOTP (authenticator app)
    openpass settings remove-totp         Disable TOTP authentication
    openpass settings show-totp-qr        Show TOTP QR code again
    openpass settings set-yubikey         Enable YubiKey authentication
    openpass settings remove-yubikey      Disable YubiKey authentication
    openpass settings help                Show this help message

DESCRIPTION:
    Configure OpenPasswd authentication methods. You can combine multiple
    methods for stronger security:
    - TOTP: Time-based codes from authenticator app
    - YubiKey: Hardware key authentication
    
    Note: Your passwords are always encrypted with your master passphrase.
    You will be prompted for your passphrase when accessing passwords.
    Storing passphrases on disk has been removed for security reasons.

EXAMPLES:
    openpass settings set-totp            # Enable Google Authenticator
    openpass settings set-yubikey         # Enable YubiKey
    openpass settings show-totp-qr        # Re-display QR code
`
	fmt.Println(help)
}

// handleMigrate handles database migrations for security improvements
func handleMigrate() {
	if len(os.Args) < 3 {
		showMigrateHelp()
		return
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "upgrade-kdf":
		handleMigrateUpgradeKDF()
	case "help", "--help", "-h":
		showMigrateHelp()
	default:
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Unknown migrate command: %s\n", subcommand)))
		showMigrateHelp()
		os.Exit(1)
	}
}

func showMigrateHelp() {
	help := `OpenPasswd - Migrate Command

COMMANDS:
    openpasswd migrate upgrade-kdf    Upgrade to stronger key derivation (600k iterations)
    openpasswd migrate help           Show this help message

DESCRIPTION:
    Migration commands to improve security of existing password databases.
    
    upgrade-kdf: Upgrades from legacy 100k iterations to current 600k iterations.
                 This makes your master passphrase harder to crack if the database
                 is stolen. All passwords will be re-encrypted.

EXAMPLES:
    openpasswd migrate upgrade-kdf    # Upgrade key derivation function

SAFETY:
    - Migrations are safe and preserve all password data
    - A backup is recommended before migrating
    - You'll need your master passphrase
`
	fmt.Println(help)
}

func handleMigrateUpgradeKDF() {
	fmt.Println(tui.ColorInfo("Upgrading KDF to 600k iterations..."))
	fmt.Println(tui.ColorWarning("This will make your passwords more secure against brute-force attacks."))
	fmt.Println()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error loading config: %v\n", err)))
		os.Exit(1)
	}

	// Check current KDF version
	if cfg.KDFVersion == crypto.KDFVersionPBKDF2_600k {
		fmt.Println(tui.ColorSuccess("Already using 600k iterations. No migration needed."))
		return
	}

	if cfg.KDFVersion > crypto.KDFVersionPBKDF2_600k {
		fmt.Println(tui.ColorSuccess(fmt.Sprintf("Already using a newer KDF version (%d). No migration needed.", cfg.KDFVersion)))
		return
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error opening database: %v\n", err)))
		os.Exit(1)
	}
	defer db.Close()

	passwords, _ := db.ListPasswords()

	if len(passwords) == 0 {
		fmt.Println(tui.ColorInfo("No passwords to migrate."))

		// Just update KDF version
		if err := config.SaveKDFVersion(crypto.KDFVersionPBKDF2_600k); err != nil {
			fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving KDF version: %v\n", err)))
			os.Exit(1)
		}

		fmt.Println(tui.ColorSuccess("✓ KDF version updated to 600k iterations"))
		return
	}

	passphrase, err := promptPassword("Enter master passphrase", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error reading passphrase: %v\n", err)))
		os.Exit(1)
	}

	// Test decryption with old KDF
	oldEncryptor := crypto.NewEncryptorWithVersion(passphrase, cfg.Salt, cfg.KDFVersion)
	testPass := passwords[0]
	if _, err := oldEncryptor.Decrypt(testPass.Name); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError("Incorrect passphrase or corrupted database\n"))
		os.Exit(1)
	}

	fmt.Println(tui.ColorInfo(fmt.Sprintf("Re-encrypting %d passwords with 600k iterations...", len(passwords))))
	fmt.Println(tui.ColorInfo("This may take a moment (stronger security = slower encryption)..."))
	fmt.Println()

	// Create new encryptor with 600k iterations
	newEncryptor := crypto.NewEncryptorWithVersion(passphrase, cfg.Salt, crypto.KDFVersionPBKDF2_600k)

	// Re-encrypt all passwords
	for i, p := range passwords {
		fmt.Printf("\rProgress: %d/%d", i+1, len(passwords))

		// Decrypt with old KDF
		if p.Name != "" {
			if decrypted, err := oldEncryptor.Decrypt(p.Name); err == nil {
				p.Name, _ = newEncryptor.Encrypt(decrypted)
			}
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
			fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("\nError updating password ID %d: %v\n", p.ID, err)))
			os.Exit(1)
		}
	}

	fmt.Println()

	// Save new KDF version
	if err := config.SaveKDFVersion(crypto.KDFVersionPBKDF2_600k); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Error saving KDF version: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("✓ Migration complete!"))
	fmt.Println(tui.ColorInfo("  KDF: PBKDF2-HMAC-SHA256 with 600,000 iterations"))
	fmt.Println(tui.ColorInfo("  Your passwords are now 6x harder to crack!"))
}

// checkForUpdatesStartup performs a non-intrusive version check on startup
func checkForUpdatesStartup() {
	// Check if version checking is disabled
	if config.IsVersionCheckDisabled() {
		return
	}

	// Use cached version check (doesn't block on network)
	latestVersion, updateAvailable, _, err := version.CheckForUpdateCached()
	if err != nil || !updateAvailable {
		return // Silently fail or no update available
	}

	// Show non-intrusive banner
	tui.RunVersionCheckBanner(version.Version, latestVersion)
}

// handleVersion displays version information
func handleVersion() {
	info := version.GetInfo()

	// Check if user wants simple version output
	hasFlag := len(os.Args) > 2

	// Simple version output (no flags)
	if !hasFlag {
		fmt.Printf("OpenPasswd v%s\n", info.Version)
		fmt.Printf("\nRun 'openpasswd version --check' to check for updates\n")
		fmt.Printf("Run 'openpasswd version --verbose' for build information\n")
		return
	}

	flag := os.Args[2]

	// Check for --verbose flag
	if flag == "--verbose" || flag == "-v" {
		fmt.Printf("OpenPasswd v%s\n", info.Version)
		fmt.Printf("\nBuild Information:\n")
		fmt.Printf("  Git Commit:  %s\n", info.GitCommit)
		fmt.Printf("  Build Date:  %s\n", info.BuildDate)
		fmt.Printf("  Go Version:  %s\n", info.GoVersion)
		fmt.Printf("  Platform:    %s\n", info.Platform)
		return
	}

	// Check for updates with interactive TUI
	if flag == "--check" || flag == "-c" {
		fmt.Println(tui.ColorInfo("Checking for updates..."))

		release, updateAvailable, err := version.CheckForUpdate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Failed to check for updates: %v\n", err)))
			return
		}

		// Prepare version info for TUI
		versionInfo := tui.VersionInfo{
			CurrentVersion:  info.Version,
			LatestVersion:   info.Version,
			UpdateAvailable: updateAvailable,
			GitCommit:       info.GitCommit,
			BuildDate:       info.BuildDate,
			GoVersion:       info.GoVersion,
			Platform:        info.Platform,
		}

		if release != nil {
			versionInfo.LatestVersion = strings.TrimPrefix(release.TagName, "v")
			versionInfo.ReleaseNotes = release.Body
			versionInfo.ReleaseURL = release.HTMLURL
		}

		// Run interactive TUI
		shouldUpgrade, err := tui.RunVersionTUI(versionInfo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("TUI error: %v\n", err)))
			return
		}

		if shouldUpgrade && updateAvailable {
			fmt.Println()
			handleUpgrade()
		}
		return
	}

	// Disable automatic version checking
	if flag == "--disable-checking" {
		if err := config.DisableVersionCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Failed to disable version checking: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println(tui.ColorSuccess("✓ Automatic version checking disabled"))
		fmt.Println(tui.ColorInfo("\nYou can still manually check for updates with:"))
		fmt.Println(tui.ColorInfo("  openpasswd version --check"))
		fmt.Println()
		fmt.Println(tui.ColorInfo("To re-enable automatic checks, run:"))
		fmt.Println(tui.ColorInfo("  openpasswd version --enable-checking"))
		return
	}

	// Enable automatic version checking
	if flag == "--enable-checking" {
		if err := config.EnableVersionCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Failed to enable version checking: %v\n", err)))
			os.Exit(1)
		}
		fmt.Println(tui.ColorSuccess("✓ Automatic version checking enabled"))
		fmt.Println(tui.ColorInfo("\nOpenPasswd will now check for updates on startup"))
		fmt.Println(tui.ColorInfo("(Updates are checked at most once per 24 hours)"))
		return
	}

	// Unknown flag
	fmt.Printf("Unknown flag: %s\n", flag)
	fmt.Printf("\nAvailable flags:\n")
	fmt.Printf("  --check, -c           Check for updates (interactive)\n")
	fmt.Printf("  --verbose, -v         Show detailed build information\n")
	fmt.Printf("  --disable-checking    Disable automatic update checks\n")
	fmt.Printf("  --enable-checking     Enable automatic update checks\n")
}

// handleUpgrade performs an automatic upgrade
func handleUpgrade() {
	fmt.Println(tui.ColorInfo("Checking for updates..."))

	release, updateAvailable, err := version.CheckForUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("Failed to check for updates: %v\n", err)))
		os.Exit(1)
	}

	if !updateAvailable {
		fmt.Printf(tui.ColorSuccess("✓ You are already running the latest version (v%s)\n"), version.Version)
		return
	}

	if release != nil {
		fmt.Printf(tui.ColorInfo("\nNew version available: v%s\n"), release.TagName)
		fmt.Printf(tui.ColorInfo("Current version: v%s\n\n"), version.Version)

		// Show release notes if available
		if release.Body != "" {
			fmt.Println(tui.ColorInfo("Release Notes:"))
			fmt.Println(tui.ColorInfo("─────────────────────────────────────────"))
			fmt.Println(release.Body)
			fmt.Println(tui.ColorInfo("─────────────────────────────────────────"))
			fmt.Println()
		}
	}

	// Confirm upgrade
	fmt.Print("Do you want to upgrade? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" && confirm != "y" {
		fmt.Println(tui.ColorInfo("Upgrade cancelled."))
		return
	}

	fmt.Println()

	// Show upgrade TUI
	newVersion := release.TagName
	if err := tui.RunUpgradeTUI(newVersion); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("TUI error: %v\n", err)))
	}

	// Actually perform the upgrade
	if err := version.Upgrade(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", tui.ColorError(fmt.Sprintf("\nUpgrade failed: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("\n✓ Upgrade completed successfully!"))
	fmt.Println(tui.ColorInfo("Please restart openpasswd to use the new version."))
}
