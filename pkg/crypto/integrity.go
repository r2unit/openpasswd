package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
)

// GenerateHMAC generates an HMAC-SHA256 signature for data
func GenerateHMAC(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// VerifyHMAC verifies an HMAC-SHA256 signature
func VerifyHMAC(data, key []byte, expectedHMAC string) bool {
	actualHMAC := GenerateHMAC(data, key)
	return hmac.Equal([]byte(actualHMAC), []byte(expectedHMAC))
}

// DeriveHMACKey derives an HMAC key from the encryption key
// This uses a separate derivation to prevent key reuse
func DeriveHMACKey(encryptionKey []byte) []byte {
	h := sha256.New()
	h.Write([]byte("openpasswd-hmac-v1"))
	h.Write(encryptionKey)
	return h.Sum(nil)
}

// SaveDatabaseHMAC generates and saves HMAC for a database file
func SaveDatabaseHMAC(dbPath string, encryptor *Encryptor) error {
	// Read database file
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database: %w", err)
	}

	// Generate HMAC
	hmacKey := DeriveHMACKey(encryptor.key)
	mac := GenerateHMAC(data, hmacKey)

	// Save HMAC to separate file
	hmacPath := dbPath + ".hmac"
	if err := os.WriteFile(hmacPath, []byte(mac), 0600); err != nil {
		return fmt.Errorf("failed to save HMAC: %w", err)
	}

	return nil
}

// VerifyDatabaseHMAC verifies the HMAC of a database file
func VerifyDatabaseHMAC(dbPath string, encryptor *Encryptor) error {
	hmacPath := dbPath + ".hmac"

	// Check if HMAC file exists
	if _, err := os.Stat(hmacPath); os.IsNotExist(err) {
		// No HMAC file means legacy database (before HMAC was added)
		// This is not an error, just return nil
		return nil
	}

	// Read database file
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database: %w", err)
	}

	// Read stored HMAC
	storedHMAC, err := os.ReadFile(hmacPath)
	if err != nil {
		return fmt.Errorf("failed to read HMAC: %w", err)
	}

	// Verify HMAC
	hmacKey := DeriveHMACKey(encryptor.key)
	if !VerifyHMAC(data, hmacKey, string(storedHMAC)) {
		return fmt.Errorf("database integrity check failed - file may be corrupted or tampered with")
	}

	return nil
}
