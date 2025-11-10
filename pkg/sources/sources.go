package sources

import (
	"github.com/r2unit/openpasswd/pkg/models"
	"github.com/r2unit/openpasswd/pkg/proton/pass"
)

// Source represents a password manager source
type Source string

const (
	SourceProtonPass Source = "protonpass"

	// SourceBitwarden  Source = "bitwarden"  // Has public API with OAuth
	// Source1Password  Source = "1password"  // Has CLI and Connect API
	// SourceLastPass   Source = "lastpass"   // CSV export only
	// SourceKeePass    Source = "keepass"    // File-based, no API
)

// Importer interface for all password manager importers
//
// This interface defines the contract for importing passwords from various
// password managers into OpenPasswd. Each importer is responsible for:
//   - Parsing the export format (JSON, CSV, etc.)
//   - Decrypting data if needed
//   - Converting to OpenPasswd's password model
//
// All password data is re-encrypted with the user's master passphrase
// before being stored in the local database.
type Importer interface {
	// Import reads passwords from a file and returns them
	// The passphrase parameter is used for encrypted exports (e.g., PGP-encrypted files)
	Import(filePath string, passphrase string) ([]*models.Password, error)

	// GetName returns the display name of this importer (e.g., "Proton Pass")
	GetName() string

	// GetDescription returns a description of this importer for UI display
	GetDescription() string

	// SupportsFormat checks if this importer supports the given file format
	// Format should include the dot (e.g., ".json", ".csv", ".zip")
	SupportsFormat(format string) bool
}

// GetImporter returns an importer for the given source
//
func GetImporter(source Source) Importer {
	switch source {
	case SourceProtonPass:
		return &pass.Importer{}
	// case SourceBitwarden:
	//     return &BitwardenImporter{}
	// case Source1Password:
	//     return &OnePasswordImporter{}
	// case SourceLastPass:
	//     return &LastPassImporter{}
	// case SourceKeePass:
	//     return &KeePassImporter{}
	default:
		return nil
	}
}

// GetAvailableImporters returns a list of all available importers
//
// This is used by the import TUI to display available password managers.
func GetAvailableImporters() []Importer {
	return []Importer{
		&pass.Importer{}, // Proton Pass (via export files only)
		// See other_importers.go for implementation notes
	}
}
