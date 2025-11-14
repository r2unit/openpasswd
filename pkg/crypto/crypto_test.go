package crypto

import (
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}

	if len(salt) != saltSize {
		t.Errorf("GenerateSalt() salt length = %d, want %d", len(salt), saltSize)
	}

	// Generate another salt and ensure it's different (extremely unlikely to be same)
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() second call error = %v", err)
	}

	if string(salt) == string(salt2) {
		t.Error("GenerateSalt() produced identical salts (should be random)")
	}
}

func TestNewEncryptor(t *testing.T) {
	passphrase := "test-passphrase"
	salt := make([]byte, saltSize)
	for i := range salt {
		salt[i] = byte(i)
	}

	enc := NewEncryptor(passphrase, salt)
	if enc == nil {
		t.Fatal("NewEncryptor() returned nil")
	}

	if len(enc.key) != keySize {
		t.Errorf("NewEncryptor() key length = %d, want %d", len(enc.key), keySize)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	passphrase := "test-passphrase-123"
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}

	enc := NewEncryptor(passphrase, salt)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "你好世界 🔐🔑"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
			"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Ciphertext should not equal plaintext
			if ciphertext == tt.plaintext {
				t.Error("Encrypt() ciphertext equals plaintext")
			}

			// Decrypt
			decrypted, err := enc.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			// Decrypted should equal original plaintext
			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	passphrase := "test-passphrase"
	salt, _ := GenerateSalt()
	enc := NewEncryptor(passphrase, salt)

	plaintext := "test message"

	// Encrypt same plaintext twice
	ciphertext1, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	ciphertext2, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Ciphertexts should be different (due to random nonce)
	if ciphertext1 == ciphertext2 {
		t.Error("Encrypt() produced same ciphertext for same plaintext (should use random nonce)")
	}

	// But both should decrypt to same plaintext
	decrypted1, _ := enc.Decrypt(ciphertext1)
	decrypted2, _ := enc.Decrypt(ciphertext2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Decrypt() failed to recover original plaintext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	salt, _ := GenerateSalt()
	enc1 := NewEncryptor("password1", salt)
	enc2 := NewEncryptor("password2", salt)

	plaintext := "secret message"
	ciphertext, err := enc1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with wrong key
	_, err = enc2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong key should fail, but succeeded")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	passphrase := "test-passphrase"
	salt, _ := GenerateSalt()
	enc := NewEncryptor(passphrase, salt)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{"empty string", ""},
		{"invalid base64", "not-valid-base64!!!"},
		{"too short", "YWJj"},             // Valid base64 but too short
		{"corrupted", "SGVsbG8gV29ybGQh"}, // Valid base64 but wrong data
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.ciphertext)
			if err == nil {
				t.Error("Decrypt() should fail for invalid ciphertext, but succeeded")
			}
		})
	}
}

func TestNewEncryptorWithVersion(t *testing.T) {
	passphrase := "test-passphrase"
	salt, _ := GenerateSalt()

	tests := []struct {
		name    string
		version int
	}{
		{"PBKDF2 100k", KDFVersionPBKDF2_100k},
		{"PBKDF2 600k", KDFVersionPBKDF2_600k},
		{"Argon2id", KDFVersionArgon2id},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewEncryptorWithVersion(passphrase, salt, tt.version)
			if enc == nil {
				t.Fatal("NewEncryptorWithVersion() returned nil")
			}

			if len(enc.key) != keySize {
				t.Errorf("NewEncryptorWithVersion() key length = %d, want %d", len(enc.key), keySize)
			}

			// Test encryption/decryption works
			plaintext := "test message"
			ciphertext, err := enc.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := enc.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestNewEncryptorArgon2id(t *testing.T) {
	passphrase := "test-passphrase"
	salt, _ := GenerateSalt()

	enc := NewEncryptorArgon2id(passphrase, salt)
	if enc == nil {
		t.Fatal("NewEncryptorArgon2id() returned nil")
	}

	if len(enc.key) != keySize {
		t.Errorf("NewEncryptorArgon2id() key length = %d, want %d", len(enc.key), keySize)
	}

	// Test encryption/decryption
	plaintext := "test message with argon2id"
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestGetKey(t *testing.T) {
	passphrase := "test-passphrase"
	salt, _ := GenerateSalt()
	enc := NewEncryptor(passphrase, salt)

	key := enc.GetKey()
	if len(key) != keySize {
		t.Errorf("GetKey() length = %d, want %d", len(key), keySize)
	}

	// Key should match internal key
	if string(key) != string(enc.key) {
		t.Error("GetKey() does not match internal key")
	}
}

func TestPBKDF2Deterministic(t *testing.T) {
	passphrase := "test-passphrase"
	salt := make([]byte, saltSize)

	// Same passphrase and salt should produce same key
	enc1 := NewEncryptor(passphrase, salt)
	enc2 := NewEncryptor(passphrase, salt)

	if string(enc1.key) != string(enc2.key) {
		t.Error("Same passphrase and salt should produce same key")
	}

	// Different salt should produce different key
	salt2, _ := GenerateSalt()
	enc3 := NewEncryptor(passphrase, salt2)

	if string(enc1.key) == string(enc3.key) {
		t.Error("Different salt should produce different key")
	}
}
