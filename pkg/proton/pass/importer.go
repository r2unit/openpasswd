package pass

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/r2unit/openpasswd/pkg/models"
)

// Importer handles imports from Proton Pass
type Importer struct{}

// Item represents an item from Proton Pass JSON export
type Item struct {
	Type        string                   `json:"type"`
	Metadata    map[string]interface{}   `json:"metadata"`
	Content     map[string]interface{}   `json:"content"`
	ExtraFields []map[string]interface{} `json:"extraFields,omitempty"`
}

// Vault represents a vault from Proton Pass JSON export
type Vault struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

// Export represents the root structure of Proton Pass JSON export
type Export struct {
	Encrypted bool    `json:"encrypted"`
	Vaults    []Vault `json:"vaults"`
}

func (p *Importer) GetName() string {
	return "Proton Pass"
}

func (p *Importer) GetDescription() string {
	return "Import passwords from Proton Pass (JSON, CSV, or encrypted export)"
}

func (p *Importer) SupportsFormat(format string) bool {
	format = strings.ToLower(format)
	return format == ".json" || format == ".csv" || format == ".zip" || format == ".pgp"
}

func (p *Importer) Import(filePath string, passphrase string) ([]*models.Password, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".zip":
		return p.importFromZip(filePath, passphrase)
	case ".json":
		return p.importFromJSON(filePath)
	case ".csv":
		return p.importFromCSV(filePath)
	case ".pgp":
		return p.importFromPGP(filePath, passphrase)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (p *Importer) importFromZip(filePath string, passphrase string) ([]*models.Password, error) {
	// Extract ZIP file
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer r.Close()

	// Look for data.json or data.pgp inside the ZIP
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "data.json") {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open file in ZIP: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read file in ZIP: %w", err)
			}

			// Parse JSON data
			return p.parseJSON(data)
		} else if strings.HasSuffix(f.Name, "data.pgp") {
			if passphrase == "" {
				return nil, fmt.Errorf("passphrase required for encrypted export")
			}

			// Extract PGP file to temp location
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open PGP file in ZIP: %w", err)
			}
			defer rc.Close()

			tmpFile, err := os.CreateTemp("", "protonpass-*.pgp")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file: %w", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, rc); err != nil {
				return nil, fmt.Errorf("failed to write temp file: %w", err)
			}
			tmpFile.Close()

			return p.importFromPGP(tmpFile.Name(), passphrase)
		}
	}

	return nil, fmt.Errorf("no supported data file found in ZIP")
}

func (p *Importer) importFromJSON(filePath string) ([]*models.Password, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	return p.parseJSON(data)
}

func (p *Importer) parseJSON(data []byte) ([]*models.Password, error) {
	var export Export
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var passwords []*models.Password

	for _, vault := range export.Vaults {
		for _, item := range vault.Items {
			password := p.convertItem(item)
			if password != nil {
				passwords = append(passwords, password)
			}
		}
	}

	return passwords, nil
}

func (p *Importer) convertItem(item Item) *models.Password {
	pwd := &models.Password{
		Fields: make(map[string]string),
	}

	// Get name from metadata
	if name, ok := item.Metadata["name"].(string); ok {
		pwd.Name = name
	}

	// Get note from metadata
	if note, ok := item.Metadata["note"].(string); ok {
		pwd.Notes = note
	}

	// Parse based on item type
	switch strings.ToLower(item.Type) {
	case "login":
		pwd.Type = models.TypeLogin
		if username, ok := item.Content["username"].(string); ok {
			pwd.Username = username
		}
		if password, ok := item.Content["password"].(string); ok {
			pwd.Password = password
		}
		if urls, ok := item.Content["urls"].([]interface{}); ok && len(urls) > 0 {
			if url, ok := urls[0].(string); ok {
				pwd.URL = url
			}
		}
		// Handle TOTP if present
		if totp, ok := item.Content["totpUri"].(string); ok {
			pwd.Fields["totp_uri"] = totp
		}

	case "note":
		pwd.Type = models.TypeNote
		if content, ok := item.Content["content"].(string); ok {
			pwd.Notes = content
		}

	case "creditcard":
		pwd.Type = models.TypeCard
		if cardholderName, ok := item.Content["cardholderName"].(string); ok {
			pwd.Fields["cardholder_name"] = cardholderName
		}
		if number, ok := item.Content["number"].(string); ok {
			pwd.Fields["number"] = number
		}
		if verificationNumber, ok := item.Content["verificationNumber"].(string); ok {
			pwd.Fields["cvv"] = verificationNumber
		}
		if expirationDate, ok := item.Content["expirationDate"].(string); ok {
			pwd.Fields["expiration_date"] = expirationDate
		}
		if pin, ok := item.Content["pin"].(string); ok {
			pwd.Fields["pin"] = pin
		}

	case "identity":
		pwd.Type = models.TypeIdentity
		// Parse identity fields
		if fullName, ok := item.Content["fullName"].(string); ok {
			pwd.Fields["full_name"] = fullName
		}
		if email, ok := item.Content["email"].(string); ok {
			pwd.Fields["email"] = email
		}
		if phoneNumber, ok := item.Content["phoneNumber"].(string); ok {
			pwd.Fields["phone_number"] = phoneNumber
		}

	default:
		pwd.Type = models.TypeOther
	}

	// Parse extra fields
	if len(item.ExtraFields) > 0 {
		for _, field := range item.ExtraFields {
			if fieldName, ok := field["fieldName"].(string); ok {
				if fieldValue, ok := field["data"].(string); ok {
					pwd.Fields[fieldName] = fieldValue
				}
			}
		}
	}

	return pwd
}

func (p *Importer) importFromCSV(filePath string) ([]*models.Password, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Parse header to find column indices
	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	var passwords []*models.Password

	for i := 1; i < len(records); i++ {
		record := records[i]
		pwd := &models.Password{
			Type:   models.TypeLogin,
			Fields: make(map[string]string),
		}

		if idx, ok := colMap["name"]; ok && idx < len(record) {
			pwd.Name = record[idx]
		}
		if idx, ok := colMap["username"]; ok && idx < len(record) {
			pwd.Username = record[idx]
		}
		if idx, ok := colMap["password"]; ok && idx < len(record) {
			pwd.Password = record[idx]
		}
		if idx, ok := colMap["url"]; ok && idx < len(record) {
			pwd.URL = record[idx]
		}
		if idx, ok := colMap["note"]; ok && idx < len(record) {
			pwd.Notes = record[idx]
		}

		passwords = append(passwords, pwd)
	}

	return passwords, nil
}

func (p *Importer) importFromPGP(filePath string, passphrase string) ([]*models.Password, error) {
	if passphrase == "" {
		return nil, fmt.Errorf("passphrase required for encrypted PGP file")
	}

	// Create temp file for decrypted output
	tmpFile, err := os.CreateTemp("", "protonpass-decrypted-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Try to decrypt using gpg
	cmd := exec.Command("gpg", "--decrypt", "--batch", "--yes", "--passphrase", passphrase, "--output", tmpFile.Name(), filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt PGP file: %s: %w", string(output), err)
	}

	// Read and parse decrypted JSON
	return p.importFromJSON(tmpFile.Name())
}
