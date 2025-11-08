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

	// All other commands require initialization
	if !isInitialized() {
		showNotInitializedError()
		os.Exit(1)
	}

	// Commands that require initialization
	switch cmd {
	// TODO: Server mode - Remote sync functionality (future feature)
	// Will allow syncing passwords across devices via a self-hosted server
	case "server":
		fmt.Fprintf(os.Stderr, tui.ColorWarning("Server mode is currently disabled\n"))
		os.Exit(1)

	// TODO: Auth - Provider authentication (future feature)
	// Will support connecting to external password providers (Proton Pass, Bitwarden, etc.)
	case "auth":
		fmt.Fprintf(os.Stderr, tui.ColorWarning("Auth functionality is currently disabled\n"))
		os.Exit(1)

	// TODO: Import - Password import from other managers (future feature)
	// Will support importing from various password manager export formats
	case "import":
		fmt.Fprintf(os.Stderr, tui.ColorWarning("Import functionality is currently disabled\n"))
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
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error generating salt: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveSalt(salt); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving configuration: %v\n", err)))
		os.Exit(1)
	}

	// Save current KDF version (600k iterations)
	if err := config.SaveKDFVersion(crypto.CurrentKDFVersion); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving KDF version: %v\n", err)))
		os.Exit(1)
	}

	// Encrypt and save recovery key
	encryptedRecovery, err := crypto.EncryptRecoveryKey(setupResult.RecoveryKey, setupResult.Passphrase, salt)
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error encrypting recovery key: %v\n", err)))
		os.Exit(1)
	}

	if err := config.SaveRecoveryKey(encryptedRecovery); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving recovery key: %v\n", err)))
		os.Exit(1)
	}

	// Save recovery key hash for verification
	recoveryHash := crypto.GenerateRecoveryHash(setupResult.RecoveryKey)
	if err := config.SaveRecoveryHash(crypto.EncodeRecoveryHash(recoveryHash)); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving recovery hash: %v\n", err)))
		os.Exit(1)
	}

	if err := config.CreateDefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorWarning(fmt.Sprintf("Warning: Could not create config.toml: %v\n", err)))
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
    openpasswd version --check                  # Check for updates
    openpasswd upgrade                          # Upgrade to latest version

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

// TODO: handleImport - Import passwords from other password managers
// This function will allow importing passwords from export files of various password managers
// Planned support: Bitwarden, 1Password, LastPass, KeePass, etc.
// Currently disabled - implementation pending
func handleImport() {
	// DISABLED: Import functionality is not yet implemented
	// Will support reading various export formats (CSV, JSON, XML)
	// and converting them to the OpenPasswd encrypted format

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
	fmt.Print("Enter master passphrase (or press Enter for none): ")
	passphrase, err := readPassword()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
		os.Exit(1)
	}
	fmt.Println()

	if err := tui.RunImportTUI(db, cfg.Salt, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

// TODO: handleAuth - Authentication with external password providers
// This function will handle OAuth/API authentication with providers like:
// - Proton Pass (via API)
// - Bitwarden (self-hosted or cloud)
// - 1Password (via CLI integration)
// Currently disabled - implementation pending
func handleAuth() {
	// DISABLED: Auth functionality is not yet implemented
	// Will require OAuth2 flows and secure token storage
	// Consider using system keychain for storing provider tokens

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

// TODO: handleAuthLogin - Login to external password provider
// Will implement OAuth2 flow or API key authentication
// Should securely store access tokens using system keychain
func handleAuthLogin() {
	// DISABLED: Not yet implemented

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

	// Always prompt for passphrase (plaintext storage removed for security)
	fmt.Print("Enter master passphrase (or press Enter for none): ")
	passphrase, err := readPassword()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
		os.Exit(1)
	}
	fmt.Println()

	if err := tui.RunAuthLoginTUI(db, cfg.Salt, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error: %v\n", err)))
		os.Exit(1)
	}
}

// TODO: handleAuthLogout - Logout from external password provider
// Will revoke access tokens and clear stored credentials
func handleAuthLogout() {
	// DISABLED: Not yet implemented
	fmt.Println(tui.ColorInfo("Auth logout functionality coming soon"))
	fmt.Println(tui.ColorInfo("For now, passwords are synced locally only"))
}

// TODO: handleAuthStatus - Check authentication status with providers
// Will display connected providers and sync status
func handleAuthStatus() {
	// DISABLED: Not yet implemented
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

	// Always prompt for passphrase (plaintext storage removed for security)
	fmt.Print("Enter master passphrase (or press Enter for none): ")
	passphrase, err := readPassword()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
		os.Exit(1)
	}
	fmt.Println()

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
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Unknown migrate command: %s\n", subcommand)))
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
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error loading config: %v\n", err)))
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
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error opening database: %v\n", err)))
		os.Exit(1)
	}
	defer db.Close()

	passwords, _ := db.ListPasswords()

	if len(passwords) == 0 {
		fmt.Println(tui.ColorInfo("No passwords to migrate."))

		// Just update KDF version
		if err := config.SaveKDFVersion(crypto.KDFVersionPBKDF2_600k); err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving KDF version: %v\n", err)))
			os.Exit(1)
		}

		fmt.Println(tui.ColorSuccess("✓ KDF version updated to 600k iterations"))
		return
	}

	fmt.Print("Enter master passphrase: ")
	passphrase, err := readPassword()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError reading passphrase: %v\n", err)))
		os.Exit(1)
	}
	fmt.Println()

	// Test decryption with old KDF
	oldEncryptor := crypto.NewEncryptorWithVersion(passphrase, cfg.Salt, cfg.KDFVersion)
	testPass := passwords[0]
	if _, err := oldEncryptor.Decrypt(testPass.Name); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError("Incorrect passphrase or corrupted database\n"))
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
			fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("\nError updating password ID %d: %v\n", p.ID, err)))
			os.Exit(1)
		}
	}

	fmt.Println()

	// Save new KDF version
	if err := config.SaveKDFVersion(crypto.KDFVersionPBKDF2_600k); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Error saving KDF version: %v\n", err)))
		os.Exit(1)
	}

	fmt.Println(tui.ColorSuccess("✓ Migration complete!"))
	fmt.Println(tui.ColorInfo("  KDF: PBKDF2-HMAC-SHA256 with 600,000 iterations"))
	fmt.Println(tui.ColorInfo("  Your passwords are now 6x harder to crack!"))
}

// handleVersion displays version information
func handleVersion() {
	info := version.GetInfo()

	// Display basic version
	fmt.Printf("OpenPasswd v%s\n", info.Version)

	// Check for --verbose flag
	if len(os.Args) > 2 && (os.Args[2] == "--verbose" || os.Args[2] == "-v") {
		fmt.Printf("\nBuild Information:\n")
		fmt.Printf("  Git Commit:  %s\n", info.GitCommit)
		fmt.Printf("  Build Date:  %s\n", info.BuildDate)
		fmt.Printf("  Go Version:  %s\n", info.GoVersion)
		fmt.Printf("  Platform:    %s\n", info.Platform)
	}

	// Check for updates
	if len(os.Args) > 2 && os.Args[2] == "--check" {
		fmt.Println("\nChecking for updates...")
		release, updateAvailable, err := version.CheckForUpdate()
		if err != nil {
			fmt.Fprintf(os.Stderr, tui.ColorWarning("Failed to check for updates: %v\n"), err)
			return
		}

		if updateAvailable && release != nil {
			fmt.Printf(tui.ColorWarning("\n⚠  Update available: v%s (you have v%s)\n"), release.TagName, info.Version)
			fmt.Printf(tui.ColorInfo("Run 'openpasswd upgrade' to update\n"))
		} else {
			fmt.Printf(tui.ColorSuccess("\n✓ You are running the latest version (v%s)\n"), info.Version)
		}
	}
}

// handleUpgrade performs an automatic upgrade
func handleUpgrade() {
	fmt.Println(tui.ColorInfo("Checking for updates..."))

	release, updateAvailable, err := version.CheckForUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Failed to check for updates: %v\n", err)))
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
	if err := version.Upgrade(); err != nil {
		fmt.Fprintf(os.Stderr, tui.ColorError(fmt.Sprintf("Upgrade failed: %v\n", err)))
		os.Exit(1)
	}
}
