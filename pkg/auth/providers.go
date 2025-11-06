package auth

import (
	"github.com/r2unit/openpasswd/pkg/models"
)

// Provider represents an authentication provider that can sync passwords
//
// This interface defines the contract for connecting to external password
// managers and syncing passwords. Each provider is responsible for:
//   - Authentication (OAuth, API tokens, file-based, etc.)
//   - Fetching passwords from the service
//   - Converting to OpenPasswd's password model
//
// Security notes:
//   - All credentials should be stored encrypted
//   - Passwords are re-encrypted with user's master passphrase
//   - Network calls should use HTTPS only
//   - Implement proper session management
type Provider interface {
	// GetName returns the display name of this provider (e.g., "Proton Pass")
	GetName() string

	// GetDescription returns a description of this provider for UI display
	GetDescription() string

	// Login authenticates with the provider
	// Credentials map contains provider-specific fields (username, password, api_token, file_path, etc.)
	// Returns error if authentication fails
	Login(credentials map[string]string) error

	// Logout logs out from the provider
	// Should clear any stored credentials/tokens
	Logout() error

	// IsAuthenticated checks if currently authenticated
	// Used to determine if we need to re-authenticate
	IsAuthenticated() bool

	// SyncPasswords syncs passwords from the provider
	// Should only be called after successful Login()
	// Returns unencrypted passwords (will be encrypted before storage)
	SyncPasswords() ([]*models.Password, error)

	// GetCredentialFields returns the fields needed for login
	// Used to dynamically generate the login form in TUI
	GetCredentialFields() []CredentialField
}

// CredentialField represents a credential field needed for login
//
// Used to dynamically generate authentication forms in the TUI.
// Different providers require different credentials:
//   - Proton Pass: file path + passphrase
//   - Bitwarden: email + password (or API token)
//   - 1Password: account + API token
type CredentialField struct {
	Name        string // Field name (e.g., "username", "password", "api_token")
	Label       string // Display label (e.g., "Username", "API Token")
	Type        string // Field type: "text", "password", "file"
	Required    bool   // Whether this field is required
	Placeholder string // Placeholder text for empty field
}

// ProviderType represents the type of provider
type ProviderType string

const (
	// Currently supported providers
	ProviderTypeProtonPass ProviderType = "protonpass" // File-based import only

	// TODO: Future providers to implement (see pkg/sources/other_importers.go)
	// ProviderTypeBitwarden  ProviderType = "bitwarden"  // OAuth + API
	// ProviderType1Password  ProviderType = "1password"  // API token
	// ProviderTypeLastPass   ProviderType = "lastpass"   // File-based only
)

// Global provider registry
// Providers auto-register themselves in their init() functions
var providers = make(map[ProviderType]Provider)

// RegisterProvider registers a new auth provider
//
// This is called automatically by each provider's init() function.
// Example from pkg/proton/pass/provider.go:
//
//	func init() {
//	    auth.RegisterProvider(auth.ProviderTypeProtonPass, &Provider{})
//	}
//
// TODO: When adding new providers, they should follow this pattern
func RegisterProvider(providerType ProviderType, provider Provider) {
	providers[providerType] = provider
}

// GetProvider returns a provider by type
//
// Returns nil if provider type is not registered.
// Check for nil before using the returned provider.
func GetProvider(providerType ProviderType) Provider {
	return providers[providerType]
}

// GetAllProviders returns all registered providers
//
// This is used by the auth login TUI to display available providers.
// Currently only returns Proton Pass provider.
//
// TODO: As more providers are implemented and registered, they will
// automatically appear in this list.
func GetAllProviders() []Provider {
	result := make([]Provider, 0, len(providers))
	for _, provider := range providers {
		result = append(result, provider)
	}
	return result
}

// GetAvailableProviders returns a list of available provider types
//
// Returns the ProviderType constants for all registered providers.
// Useful for validation or config file generation.
func GetAvailableProviders() []ProviderType {
	result := make([]ProviderType, 0, len(providers))
	for providerType := range providers {
		result = append(result, providerType)
	}
	return result
}
