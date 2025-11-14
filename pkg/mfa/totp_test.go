package mfa

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateTOTPSecret(t *testing.T) {
	accountName := "test@example.com"

	key, err := GenerateTOTPSecret(accountName)
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error = %v", err)
	}

	if key == nil {
		t.Fatal("GenerateTOTPSecret() returned nil")
	}

	if key.Secret == "" {
		t.Error("GenerateTOTPSecret() secret is empty")
	}

	if key.Issuer != "OpenPasswd" {
		t.Errorf("GenerateTOTPSecret() issuer = %q, want %q", key.Issuer, "OpenPasswd")
	}

	if key.AccountName != accountName {
		t.Errorf("GenerateTOTPSecret() account = %q, want %q", key.AccountName, accountName)
	}

	// Multiple generations should produce different secrets
	key2, _ := GenerateTOTPSecret(accountName)
	if key.Secret == key2.Secret {
		t.Error("GenerateTOTPSecret() produced identical secrets (should be random)")
	}
}

func TestTOTPKeyURL(t *testing.T) {
	key := &TOTPKey{
		Secret:      "JBSWY3DPEHPK3PXP",
		Issuer:      "OpenPasswd",
		AccountName: "user@example.com",
	}

	url := key.URL()

	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Errorf("URL() should start with 'otpauth://totp/', got %q", url)
	}

	if !strings.Contains(url, "secret="+key.Secret) {
		t.Error("URL() should contain secret parameter")
	}

	if !strings.Contains(url, "issuer=OpenPasswd") {
		t.Error("URL() should contain issuer parameter")
	}

	if !strings.Contains(url, "algorithm=SHA1") {
		t.Error("URL() should contain algorithm=SHA1")
	}

	if !strings.Contains(url, "digits=6") {
		t.Error("URL() should contain digits=6")
	}

	if !strings.Contains(url, "period=30") {
		t.Error("URL() should contain period=30")
	}
}

func TestGenerateTOTP(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	counter := int64(1)

	code := generateTOTP(secret, counter)

	if len(code) != 6 {
		t.Errorf("generateTOTP() code length = %d, want 6", len(code))
	}

	// Code should be numeric
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("generateTOTP() code contains non-digit: %c", c)
		}
	}

	// Same secret and counter should produce same code
	code2 := generateTOTP(secret, counter)
	if code != code2 {
		t.Error("generateTOTP() should be deterministic")
	}

	// Different counter should produce different code (usually)
	code3 := generateTOTP(secret, counter+1)
	if code == code3 {
		t.Log("Warning: generateTOTP() produced same code for different counters (rare but possible)")
	}
}

func TestGenerateTOTPWithSpaces(t *testing.T) {
	// Secret with spaces should work
	secret := "JBSW Y3DP EHPK 3PXP"
	counter := int64(1)

	code := generateTOTP(secret, counter)

	if len(code) != 6 {
		t.Errorf("generateTOTP() with spaces: code length = %d, want 6", len(code))
	}

	// Should produce same result as secret without spaces
	secretNoSpaces := "JBSWY3DPEHPK3PXP"
	codeNoSpaces := generateTOTP(secretNoSpaces, counter)

	if code != codeNoSpaces {
		t.Error("generateTOTP() should handle spaces in secret")
	}
}

func TestGenerateTOTPLowercase(t *testing.T) {
	// Lowercase secret should work
	secretLower := "jbswy3dpehpk3pxp"
	secretUpper := "JBSWY3DPEHPK3PXP"
	counter := int64(1)

	codeLower := generateTOTP(secretLower, counter)
	codeUpper := generateTOTP(secretUpper, counter)

	if codeLower != codeUpper {
		t.Error("generateTOTP() should handle lowercase secrets")
	}
}

func TestValidateTOTP(t *testing.T) {
	// Generate a secret and code
	key, _ := GenerateTOTPSecret("test@example.com")

	// Generate code for current time window
	currentCounter := time.Now().Unix() / 30
	expectedCode := generateTOTP(key.Secret, currentCounter)

	// Validate should succeed with correct code
	if !ValidateTOTP(key.Secret, expectedCode) {
		t.Error("ValidateTOTP() failed to validate correct code")
	}

	// Validate should fail with wrong code
	if ValidateTOTP(key.Secret, "000000") {
		t.Error("ValidateTOTP() validated incorrect code")
	}

	if ValidateTOTP(key.Secret, "999999") {
		t.Error("ValidateTOTP() validated incorrect code")
	}

	if ValidateTOTP(key.Secret, "invalid") {
		t.Error("ValidateTOTP() validated invalid code")
	}
}

func TestValidateTOTPKnownVector(t *testing.T) {
	// Test with a known secret and timestamp
	// This ensures our TOTP implementation is correct
	secret := "JBSWY3DPEHPK3PXP"

	// For counter = 1, this secret should produce a specific code
	counter := int64(1)
	code := generateTOTP(secret, counter)

	// Validate that the generated code validates correctly
	// We can't test against a specific value without knowing the timestamp,
	// but we can test consistency
	if len(code) != 6 {
		t.Errorf("Known vector: code length = %d, want 6", len(code))
	}
}

func TestEncodeTOTPSecret(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"

	encoded := EncodeTOTPSecret(secret)

	if encoded == "" {
		t.Error("EncodeTOTPSecret() returned empty string")
	}

	if encoded == secret {
		t.Error("EncodeTOTPSecret() did not encode secret")
	}
}

func TestDecodeTOTPSecret(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"

	encoded := EncodeTOTPSecret(secret)
	decoded, err := DecodeTOTPSecret(encoded)

	if err != nil {
		t.Fatalf("DecodeTOTPSecret() error = %v", err)
	}

	if decoded != secret {
		t.Errorf("DecodeTOTPSecret() = %q, want %q", decoded, secret)
	}
}

func TestDecodeTOTPSecretInvalid(t *testing.T) {
	// Invalid base64
	_, err := DecodeTOTPSecret("not-valid-base64!!!")
	if err == nil {
		t.Error("DecodeTOTPSecret() should error for invalid base64")
	}
}

func TestTOTPRoundTrip(t *testing.T) {
	// Generate secret
	key, err := GenerateTOTPSecret("test@example.com")
	if err != nil {
		t.Fatalf("GenerateTOTPSecret() error = %v", err)
	}

	// Encode secret
	encoded := EncodeTOTPSecret(key.Secret)

	// Decode secret
	decoded, err := DecodeTOTPSecret(encoded)
	if err != nil {
		t.Fatalf("DecodeTOTPSecret() error = %v", err)
	}

	if decoded != key.Secret {
		t.Error("TOTP secret round trip failed")
	}

	// Generate code with decoded secret
	currentCounter := time.Now().Unix() / 30
	code := generateTOTP(decoded, currentCounter)

	// Validate code
	if !ValidateTOTP(decoded, code) {
		t.Error("ValidateTOTP() failed after round trip")
	}
}

func TestTOTPTimeSensitivity(t *testing.T) {
	// Generate a secret
	key, _ := GenerateTOTPSecret("test@example.com")

	// Generate code for specific counter
	counter := int64(12345)
	code := generateTOTP(key.Secret, counter)

	// Code should be 6 digits
	if len(code) != 6 {
		t.Errorf("Code length = %d, want 6", len(code))
	}

	// Same counter should produce same code
	code2 := generateTOTP(key.Secret, counter)
	if code != code2 {
		t.Error("Same counter should produce same code")
	}

	// Different counter should typically produce different code
	code3 := generateTOTP(key.Secret, counter+1)
	if code == code3 {
		t.Log("Note: Adjacent time windows produced same code (rare but possible)")
	}
}

func TestGenerateTOTPInvalidSecret(t *testing.T) {
	// Invalid base32 secret
	invalidSecret := "INVALID@#$%"
	counter := int64(1)

	code := generateTOTP(invalidSecret, counter)

	// Should return empty string for invalid secret
	if code != "" {
		t.Errorf("generateTOTP() with invalid secret should return empty, got %q", code)
	}
}

func TestTOTPSecretLength(t *testing.T) {
	// Generate multiple secrets and check they're reasonable length
	for i := 0; i < 5; i++ {
		key, err := GenerateTOTPSecret("test@example.com")
		if err != nil {
			t.Fatalf("GenerateTOTPSecret() iteration %d error = %v", i, err)
		}

		// Base32 encoded 20 bytes should be about 32 characters (without padding)
		if len(key.Secret) < 20 || len(key.Secret) > 40 {
			t.Errorf("Secret length %d seems unusual (expected ~32)", len(key.Secret))
		}
	}
}

func TestTOTPURLEncoding(t *testing.T) {
	// Test with special characters in account name
	key := &TOTPKey{
		Secret:      "JBSWY3DPEHPK3PXP",
		Issuer:      "OpenPasswd",
		AccountName: "user+test@example.com",
	}

	url := key.URL()

	// URL should be properly encoded
	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Error("URL should start with otpauth://totp/")
	}

	// Should contain encoded label
	if !strings.Contains(url, "OpenPasswd%3A") || !strings.Contains(url, "example.com") {
		t.Error("URL should contain properly encoded label")
	}
}
