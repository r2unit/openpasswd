package pass

import (
	"fmt"

	"github.com/r2unit/openpasswd/pkg/auth"
	"github.com/r2unit/openpasswd/pkg/models"
)

// Provider implements the auth.Provider interface for Proton Pass
//
// IMPORTANT: Proton Pass does NOT have a public API for third-party integrations.
// This provider works by importing from Proton Pass export files only.
//
// Supported export formats:
//   - JSON (unencrypted)
//   - CSV (simple format)
//   - ZIP (containing JSON or PGP-encrypted data)
//   - PGP (encrypted, requires gpg installed)
//
// How to use:
//  1. Export from Proton Pass: Settings → Export
//  2. Choose export format (PGP-encrypted ZIP recommended)
//  3. Run: openpasswd auth login
//  4. Select "Proton Pass" and provide file path + passphrase
//
// This is NOT a live sync - it's a one-time import from export files.
// This is the only officially supported method from Proton.
//
// TODO: If Proton releases a public API in the future, we could implement
// live sync here. Monitor: https://github.com/ProtonMail/go-proton-api
type Provider struct {
	importer      *Importer // Handles parsing export files
	authenticated bool      // Set to true after successful file validation
	filePath      string    // Path to export file
	passphrase    string    // Passphrase for encrypted exports (optional)
}

func init() {
	// Auto-register Proton Pass provider on package initialization
	// This makes it available in the auth login menu automatically
	auth.RegisterProvider(auth.ProviderTypeProtonPass, &Provider{
		importer: &Importer{},
	})
}

func (p *Provider) GetName() string {
	return "Proton Pass"
}

func (p *Provider) GetDescription() string {
	return "Import passwords from Proton Pass export file (JSON, CSV, or encrypted)"
}

func (p *Provider) GetCredentialFields() []auth.CredentialField {
	return []auth.CredentialField{
		{
			Name:        "file_path",
			Label:       "Export File Path",
			Type:        "file",
			Required:    true,
			Placeholder: "/path/to/proton-pass-export.zip",
		},
		{
			Name:        "passphrase",
			Label:       "Export Passphrase",
			Type:        "password",
			Required:    false,
			Placeholder: "Leave empty for unencrypted exports",
		},
	}
}

func (p *Provider) Login(credentials map[string]string) error {
	filePath, ok := credentials["file_path"]
	if !ok || filePath == "" {
		return fmt.Errorf("file path is required")
	}

	passphrase := credentials["passphrase"]

	p.filePath = filePath
	p.passphrase = passphrase
	p.authenticated = true

	return nil
}

func (p *Provider) Logout() error {
	p.authenticated = false
	p.filePath = ""
	p.passphrase = ""
	return nil
}

func (p *Provider) IsAuthenticated() bool {
	return p.authenticated
}

func (p *Provider) SyncPasswords() ([]*models.Password, error) {
	if !p.authenticated {
		return nil, fmt.Errorf("not authenticated")
	}

	passwords, err := p.importer.Import(p.filePath, p.passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to import from Proton Pass: %w", err)
	}

	return passwords, nil
}
