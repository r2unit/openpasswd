package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateHMAC(t *testing.T) {
	data := []byte("test data")
	key := []byte("test key")

	hmac := GenerateHMAC(data, key)

	if hmac == "" {
		t.Error("GenerateHMAC() returned empty string")
	}

	// Test determinism
	hmac2 := GenerateHMAC(data, key)
	if hmac != hmac2 {
		t.Error("GenerateHMAC() should be deterministic")
	}
}

func TestVerifyHMAC(t *testing.T) {
	data := []byte("test data")
	key := []byte("test key")

	hmac := GenerateHMAC(data, key)

	// Valid HMAC should verify
	if !VerifyHMAC(data, key, hmac) {
		t.Error("VerifyHMAC() failed to verify valid HMAC")
	}

	// Modified data should fail
	if VerifyHMAC([]byte("modified data"), key, hmac) {
		t.Error("VerifyHMAC() verified HMAC with modified data")
	}

	// Wrong key should fail
	if VerifyHMAC(data, []byte("wrong key"), hmac) {
		t.Error("VerifyHMAC() verified HMAC with wrong key")
	}

	// Invalid HMAC should fail
	if VerifyHMAC(data, key, "invalid-hmac") {
		t.Error("VerifyHMAC() verified invalid HMAC")
	}
}

func TestDeriveHMACKey(t *testing.T) {
	encKey := []byte("encryption-key-12345678901234567890")

	hmacKey := DeriveHMACKey(encKey)

	if len(hmacKey) != 32 { // SHA256 output
		t.Errorf("DeriveHMACKey() length = %d, want 32", len(hmacKey))
	}

	// Test determinism
	hmacKey2 := DeriveHMACKey(encKey)
	if string(hmacKey) != string(hmacKey2) {
		t.Error("DeriveHMACKey() should be deterministic")
	}

	// Different encryption key should produce different HMAC key
	encKey2 := []byte("different-key-12345678901234567890")
	hmacKey3 := DeriveHMACKey(encKey2)
	if string(hmacKey) == string(hmacKey3) {
		t.Error("DeriveHMACKey() should produce different keys for different inputs")
	}
}

func TestSaveAndVerifyDatabaseHMAC(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create test database file
	testData := []byte(`{"passwords": [], "next_id": 1}`)
	if err := os.WriteFile(dbPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create encryptor
	salt, _ := GenerateSalt()
	enc := NewEncryptor("test-passphrase", salt)

	// Save HMAC
	if err := SaveDatabaseHMAC(dbPath, enc); err != nil {
		t.Fatalf("SaveDatabaseHMAC() error = %v", err)
	}

	// Verify HMAC file was created
	hmacPath := dbPath + ".hmac"
	if _, err := os.Stat(hmacPath); os.IsNotExist(err) {
		t.Error("SaveDatabaseHMAC() did not create HMAC file")
	}

	// Verify HMAC
	if err := VerifyDatabaseHMAC(dbPath, enc); err != nil {
		t.Errorf("VerifyDatabaseHMAC() error = %v", err)
	}
}

func TestVerifyDatabaseHMACWithoutFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create test database file
	testData := []byte(`{"passwords": [], "next_id": 1}`)
	if err := os.WriteFile(dbPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create encryptor
	salt, _ := GenerateSalt()
	enc := NewEncryptor("test-passphrase", salt)

	// Verify without HMAC file (should not error for legacy databases)
	if err := VerifyDatabaseHMAC(dbPath, enc); err != nil {
		t.Errorf("VerifyDatabaseHMAC() should not error when HMAC file doesn't exist (legacy): %v", err)
	}
}

func TestVerifyDatabaseHMACTamperedData(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create test database file
	testData := []byte(`{"passwords": [], "next_id": 1}`)
	if err := os.WriteFile(dbPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create encryptor
	salt, _ := GenerateSalt()
	enc := NewEncryptor("test-passphrase", salt)

	// Save HMAC
	if err := SaveDatabaseHMAC(dbPath, enc); err != nil {
		t.Fatalf("SaveDatabaseHMAC() error = %v", err)
	}

	// Tamper with database
	tamperedData := []byte(`{"passwords": [], "next_id": 999}`)
	if err := os.WriteFile(dbPath, tamperedData, 0600); err != nil {
		t.Fatalf("Failed to tamper with database: %v", err)
	}

	// Verify should fail
	if err := VerifyDatabaseHMAC(dbPath, enc); err == nil {
		t.Error("VerifyDatabaseHMAC() should fail for tampered database")
	}
}

func TestVerifyDatabaseHMACWrongKey(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create test database file
	testData := []byte(`{"passwords": [], "next_id": 1}`)
	if err := os.WriteFile(dbPath, testData, 0600); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create encryptor
	salt, _ := GenerateSalt()
	enc1 := NewEncryptor("test-passphrase", salt)

	// Save HMAC
	if err := SaveDatabaseHMAC(dbPath, enc1); err != nil {
		t.Fatalf("SaveDatabaseHMAC() error = %v", err)
	}

	// Verify with different key
	enc2 := NewEncryptor("wrong-passphrase", salt)
	if err := VerifyDatabaseHMAC(dbPath, enc2); err == nil {
		t.Error("VerifyDatabaseHMAC() should fail with wrong key")
	}
}

func TestSaveDatabaseHMACNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent.db")

	salt, _ := GenerateSalt()
	enc := NewEncryptor("test-passphrase", salt)

	// Should error for non-existent file
	if err := SaveDatabaseHMAC(dbPath, enc); err == nil {
		t.Error("SaveDatabaseHMAC() should error for non-existent file")
	}
}

func TestVerifyDatabaseHMACNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nonexistent.db")

	salt, _ := GenerateSalt()
	enc := NewEncryptor("test-passphrase", salt)

	// Should error for non-existent file
	if err := VerifyDatabaseHMAC(dbPath, enc); err == nil {
		t.Error("VerifyDatabaseHMAC() should error for non-existent database file")
	}
}
