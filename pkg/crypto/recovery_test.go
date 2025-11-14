package crypto

import (
	"strings"
	"testing"
)

func TestGenerateRecoveryKey(t *testing.T) {
	key, err := GenerateRecoveryKey()
	if err != nil {
		t.Fatalf("GenerateRecoveryKey() error = %v", err)
	}

	// Should have 24 words separated by hyphens
	words := strings.Split(key, "-")
	if len(words) != 24 {
		t.Errorf("GenerateRecoveryKey() produced %d words, want 24", len(words))
	}

	// Each word should be from the word list
	for i, word := range words {
		found := false
		for _, validWord := range recoveryWords {
			if word == validWord {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Word %d (%q) is not in recovery word list", i, word)
		}
	}

	// Multiple generations should produce different keys
	key2, _ := GenerateRecoveryKey()
	if key == key2 {
		t.Error("GenerateRecoveryKey() produced identical keys (should be random)")
	}
}

func TestRecoveryKeyToSeed(t *testing.T) {
	// Generate a key and convert to seed
	key, err := GenerateRecoveryKey()
	if err != nil {
		t.Fatalf("GenerateRecoveryKey() error = %v", err)
	}

	seed, err := RecoveryKeyToSeed(key)
	if err != nil {
		t.Fatalf("RecoveryKeyToSeed() error = %v", err)
	}

	if len(seed) != 32 {
		t.Errorf("RecoveryKeyToSeed() seed length = %d, want 32", len(seed))
	}

	// Same key should produce same seed
	seed2, _ := RecoveryKeyToSeed(key)
	if string(seed) != string(seed2) {
		t.Error("RecoveryKeyToSeed() should be deterministic")
	}
}

func TestRecoveryKeyToSeedInvalid(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"too few words", "word1-word2-word3"},
		{"too many words", strings.Repeat("word-", 30)},
		{"invalid word", strings.Repeat("invalidword-", 23) + "invalidword"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RecoveryKeyToSeed(tt.key)
			if err == nil {
				t.Error("RecoveryKeyToSeed() should error for invalid key")
			}
		})
	}
}

func TestDerivePassphraseFromRecovery(t *testing.T) {
	key, _ := GenerateRecoveryKey()
	salt, _ := GenerateSalt()

	passphrase, err := DerivePassphraseFromRecovery(key, salt)
	if err != nil {
		t.Fatalf("DerivePassphraseFromRecovery() error = %v", err)
	}

	if len(passphrase) != 32 {
		t.Errorf("DerivePassphraseFromRecovery() length = %d, want 32", len(passphrase))
	}

	// Same key and salt should produce same passphrase
	passphrase2, _ := DerivePassphraseFromRecovery(key, salt)
	if string(passphrase) != string(passphrase2) {
		t.Error("DerivePassphraseFromRecovery() should be deterministic")
	}

	// Different salt should produce different passphrase
	salt2, _ := GenerateSalt()
	passphrase3, _ := DerivePassphraseFromRecovery(key, salt2)
	if string(passphrase) == string(passphrase3) {
		t.Error("DerivePassphraseFromRecovery() with different salt should produce different passphrase")
	}
}

func TestEncryptDecryptRecoveryKey(t *testing.T) {
	recoveryKey, _ := GenerateRecoveryKey()
	passphrase := "test-passphrase"
	salt, _ := GenerateSalt()

	// Encrypt
	encrypted, err := EncryptRecoveryKey(recoveryKey, passphrase, salt)
	if err != nil {
		t.Fatalf("EncryptRecoveryKey() error = %v", err)
	}

	if encrypted == recoveryKey {
		t.Error("EncryptRecoveryKey() did not encrypt")
	}

	// Decrypt
	decrypted, err := DecryptRecoveryKey(encrypted, passphrase, salt)
	if err != nil {
		t.Fatalf("DecryptRecoveryKey() error = %v", err)
	}

	if decrypted != recoveryKey {
		t.Errorf("DecryptRecoveryKey() = %q, want %q", decrypted, recoveryKey)
	}
}

func TestDecryptRecoveryKeyWrongPassphrase(t *testing.T) {
	recoveryKey, _ := GenerateRecoveryKey()
	salt, _ := GenerateSalt()

	encrypted, _ := EncryptRecoveryKey(recoveryKey, "correct-passphrase", salt)

	// Try to decrypt with wrong passphrase
	_, err := DecryptRecoveryKey(encrypted, "wrong-passphrase", salt)
	if err == nil {
		t.Error("DecryptRecoveryKey() with wrong passphrase should fail")
	}
}

func TestFormatRecoveryKey(t *testing.T) {
	key, _ := GenerateRecoveryKey()
	formatted := FormatRecoveryKey(key)

	if formatted == "" {
		t.Error("FormatRecoveryKey() returned empty string")
	}

	// Should contain numbered lines
	if !strings.Contains(formatted, "1.") {
		t.Error("FormatRecoveryKey() should contain line numbers")
	}

	// Test with invalid key (not 24 words)
	invalidKey := "word1-word2-word3"
	formatted = FormatRecoveryKey(invalidKey)
	if formatted != invalidKey {
		t.Error("FormatRecoveryKey() should return key as-is for invalid format")
	}
}

func TestGenerateRecoveryHash(t *testing.T) {
	key, _ := GenerateRecoveryKey()
	hash := GenerateRecoveryHash(key)

	if len(hash) != 32 {
		t.Errorf("GenerateRecoveryHash() length = %d, want 32", len(hash))
	}

	// Same key should produce same hash
	hash2 := GenerateRecoveryHash(key)
	if string(hash) != string(hash2) {
		t.Error("GenerateRecoveryHash() should be deterministic")
	}

	// Different key should produce different hash
	key2, _ := GenerateRecoveryKey()
	hash3 := GenerateRecoveryHash(key2)
	if string(hash) == string(hash3) {
		t.Error("GenerateRecoveryHash() should produce different hashes for different keys")
	}
}

func TestVerifyRecoveryKey(t *testing.T) {
	key, _ := GenerateRecoveryKey()
	hash := GenerateRecoveryHash(key)

	// Correct key should verify
	if !VerifyRecoveryKey(key, hash) {
		t.Error("VerifyRecoveryKey() failed to verify correct key")
	}

	// Wrong key should not verify
	key2, _ := GenerateRecoveryKey()
	if VerifyRecoveryKey(key2, hash) {
		t.Error("VerifyRecoveryKey() verified wrong key")
	}

	// Invalid hash length should not verify
	if VerifyRecoveryKey(key, []byte("short")) {
		t.Error("VerifyRecoveryKey() verified with invalid hash length")
	}
}

func TestEncodeDecodeRecoveryHash(t *testing.T) {
	key, _ := GenerateRecoveryKey()
	hash := GenerateRecoveryHash(key)

	// Encode
	encoded := EncodeRecoveryHash(hash)
	if encoded == "" {
		t.Error("EncodeRecoveryHash() returned empty string")
	}

	// Decode
	decoded, err := DecodeRecoveryHash(encoded)
	if err != nil {
		t.Fatalf("DecodeRecoveryHash() error = %v", err)
	}

	if string(decoded) != string(hash) {
		t.Error("DecodeRecoveryHash() did not match original hash")
	}
}

func TestDecodeRecoveryHashInvalid(t *testing.T) {
	_, err := DecodeRecoveryHash("not-valid-base64!!!")
	if err == nil {
		t.Error("DecodeRecoveryHash() should error for invalid base64")
	}
}

func TestGenerateRecoveryKeyWithChecksum(t *testing.T) {
	keyWithChecksum, err := GenerateRecoveryKeyWithChecksum()
	if err != nil {
		t.Fatalf("GenerateRecoveryKeyWithChecksum() error = %v", err)
	}

	// Should have 25 words (24 + checksum)
	words := strings.Split(keyWithChecksum, "-")
	if len(words) != 25 {
		t.Errorf("GenerateRecoveryKeyWithChecksum() produced %d words, want 25", len(words))
	}

	// Should verify its own checksum
	valid, key := VerifyRecoveryKeyChecksum(keyWithChecksum)
	if !valid {
		t.Error("GenerateRecoveryKeyWithChecksum() produced key with invalid checksum")
	}

	// Key without checksum should be 24 words
	keyWords := strings.Split(key, "-")
	if len(keyWords) != 24 {
		t.Errorf("Extracted key has %d words, want 24", len(keyWords))
	}
}

func TestVerifyRecoveryKeyChecksum(t *testing.T) {
	// Generate key with checksum
	keyWithChecksum, _ := GenerateRecoveryKeyWithChecksum()

	// Should verify
	valid, key := VerifyRecoveryKeyChecksum(keyWithChecksum)
	if !valid {
		t.Error("VerifyRecoveryKeyChecksum() failed to verify valid checksum")
	}

	if key == "" {
		t.Error("VerifyRecoveryKeyChecksum() returned empty key")
	}

	// Corrupt checksum
	parts := strings.Split(keyWithChecksum, "-")
	parts[24] = "wrongword"
	corruptedKey := strings.Join(parts, "-")

	valid, _ = VerifyRecoveryKeyChecksum(corruptedKey)
	if valid {
		t.Error("VerifyRecoveryKeyChecksum() verified corrupted checksum")
	}

	// Wrong number of words
	valid, _ = VerifyRecoveryKeyChecksum("word1-word2-word3")
	if valid {
		t.Error("VerifyRecoveryKeyChecksum() verified key with wrong word count")
	}
}

func TestRecoveryKeyRoundTrip(t *testing.T) {
	// Generate recovery key
	recoveryKey, _ := GenerateRecoveryKey()

	// Convert to seed
	seed, _ := RecoveryKeyToSeed(recoveryKey)

	// Use seed to derive passphrase
	salt, _ := GenerateSalt()
	passphrase, _ := DerivePassphraseFromRecovery(recoveryKey, salt)

	// Verify passphrase can be used for encryption
	enc := NewEncryptor(string(passphrase), salt)
	plaintext := "test message"
	ciphertext, _ := enc.Encrypt(plaintext)
	decrypted, _ := enc.Decrypt(ciphertext)

	if decrypted != plaintext {
		t.Error("Recovery key round trip failed: decrypted text doesn't match")
	}

	// Verify seed is deterministic
	seed2, _ := RecoveryKeyToSeed(recoveryKey)
	if string(seed) != string(seed2) {
		t.Error("RecoveryKeyToSeed() should be deterministic across calls")
	}
}
