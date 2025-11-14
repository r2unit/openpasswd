package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}

	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Should contain .config/openpasswd
	if !filepath.IsAbs(dir) {
		t.Error("GetConfigDir() should return absolute path")
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	dir, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	// Directory should exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("EnsureConfigDir() did not create directory")
	}

	// Verify permissions (should be 0700)
	info, _ := os.Stat(dir)
	mode := info.Mode().Perm()
	if mode != 0700 {
		t.Errorf("EnsureConfigDir() mode = %o, want 0700", mode)
	}
}

func TestDefaultKeybindings(t *testing.T) {
	kb := DefaultKeybindings()

	tests := []struct {
		name  string
		value string
	}{
		{"Quit", kb.Quit},
		{"QuitAlt", kb.QuitAlt},
		{"Back", kb.Back},
		{"Up", kb.Up},
		{"UpAlt", kb.UpAlt},
		{"Down", kb.Down},
		{"DownAlt", kb.DownAlt},
		{"Select", kb.Select},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("DefaultKeybindings() %s is empty", tt.name)
			}
		})
	}

	// Check specific expected values
	if kb.Quit != ":q" {
		t.Errorf("DefaultKeybindings() Quit = %q, want %q", kb.Quit, ":q")
	}

	if kb.Select != "enter" {
		t.Errorf("DefaultKeybindings() Select = %q, want %q", kb.Select, "enter")
	}
}

func TestSaveAndLoadSalt(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Generate and save salt
	salt := []byte("test-salt-32-bytes-long-test!")
	if len(salt) < 32 {
		t.Fatal("Test salt too short")
	}

	err := SaveSalt(salt)
	if err != nil {
		t.Fatalf("SaveSalt() error = %v", err)
	}

	// Load config (which loads salt)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify salt matches
	if string(cfg.Salt) != string(salt) {
		t.Error("Loaded salt does not match saved salt")
	}
}

func TestLoadConfigWithoutInit(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Try to load config without initializing
	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should error when not initialized")
	}
}

func TestSaveAndLoadKDFVersion(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Save version
	version := 3
	err := SaveKDFVersion(version)
	if err != nil {
		t.Fatalf("SaveKDFVersion() error = %v", err)
	}

	// Load version
	loaded, err := LoadKDFVersion()
	if err != nil {
		t.Fatalf("LoadKDFVersion() error = %v", err)
	}

	if loaded != version {
		t.Errorf("LoadKDFVersion() = %d, want %d", loaded, version)
	}
}

func TestLoadKDFVersionDefault(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Load without saving (should default to 1)
	version, _ := LoadKDFVersion()
	if version != 1 {
		t.Errorf("LoadKDFVersion() default = %d, want 1", version)
	}
}

func TestLoadKDFVersionInvalid(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Create invalid version file
	dir, _ := EnsureConfigDir()
	versionPath := filepath.Join(dir, "kdf_version")
	os.WriteFile(versionPath, []byte("999"), 0600) // Invalid version

	// Should default to 1
	version, _ := LoadKDFVersion()
	if version != 1 {
		t.Errorf("LoadKDFVersion() with invalid version = %d, want 1", version)
	}
}

func TestHasTOTP(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Should be false initially
	if HasTOTP() {
		t.Error("HasTOTP() should be false before saving TOTP")
	}

	// Save TOTP secret
	err := SaveTOTPSecret("test-secret")
	if err != nil {
		t.Fatalf("SaveTOTPSecret() error = %v", err)
	}

	// Should be true after saving
	if !HasTOTP() {
		t.Error("HasTOTP() should be true after saving TOTP")
	}
}

func TestSaveAndLoadTOTPSecret(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	secret := "JBSWY3DPEHPK3PXP"

	// Save secret
	err := SaveTOTPSecret(secret)
	if err != nil {
		t.Fatalf("SaveTOTPSecret() error = %v", err)
	}

	// Load secret
	loaded, err := LoadTOTPSecret()
	if err != nil {
		t.Fatalf("LoadTOTPSecret() error = %v", err)
	}

	if loaded != secret {
		t.Errorf("LoadTOTPSecret() = %q, want %q", loaded, secret)
	}
}

func TestRemoveTOTP(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Save TOTP
	SaveTOTPSecret("test-secret")

	// Remove TOTP
	err := RemoveTOTP()
	if err != nil {
		t.Fatalf("RemoveTOTP() error = %v", err)
	}

	// Should not have TOTP anymore
	if HasTOTP() {
		t.Error("HasTOTP() should be false after removal")
	}

	// Removing again should not error
	err = RemoveTOTP()
	if err != nil {
		t.Error("RemoveTOTP() should not error when file doesn't exist")
	}
}

func TestHasYubiKey(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Should be false initially
	if HasYubiKey() {
		t.Error("HasYubiKey() should be false before saving")
	}

	// Save challenge
	err := SaveYubiKeyChallenge("test-challenge")
	if err != nil {
		t.Fatalf("SaveYubiKeyChallenge() error = %v", err)
	}

	// Should be true after saving
	if !HasYubiKey() {
		t.Error("HasYubiKey() should be true after saving")
	}
}

func TestSaveAndLoadYubiKeyChallenge(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	challenge := "test-challenge-string"

	// Save challenge
	err := SaveYubiKeyChallenge(challenge)
	if err != nil {
		t.Fatalf("SaveYubiKeyChallenge() error = %v", err)
	}

	// Load challenge
	loaded, err := LoadYubiKeyChallenge()
	if err != nil {
		t.Fatalf("LoadYubiKeyChallenge() error = %v", err)
	}

	if loaded != challenge {
		t.Errorf("LoadYubiKeyChallenge() = %q, want %q", loaded, challenge)
	}
}

func TestRemoveYubiKey(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Save YubiKey
	SaveYubiKeyChallenge("test-challenge")

	// Remove YubiKey
	err := RemoveYubiKey()
	if err != nil {
		t.Fatalf("RemoveYubiKey() error = %v", err)
	}

	// Should not have YubiKey anymore
	if HasYubiKey() {
		t.Error("HasYubiKey() should be false after removal")
	}
}

func TestHasRecoveryKey(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Should be false initially
	if HasRecoveryKey() {
		t.Error("HasRecoveryKey() should be false before saving")
	}

	// Save recovery key
	err := SaveRecoveryKey("encrypted-recovery-key")
	if err != nil {
		t.Fatalf("SaveRecoveryKey() error = %v", err)
	}

	// Should be true after saving
	if !HasRecoveryKey() {
		t.Error("HasRecoveryKey() should be true after saving")
	}
}

func TestSaveAndLoadRecoveryKey(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	encryptedKey := "encrypted-recovery-key-data"

	// Save recovery key
	err := SaveRecoveryKey(encryptedKey)
	if err != nil {
		t.Fatalf("SaveRecoveryKey() error = %v", err)
	}

	// Load recovery key
	loaded, err := LoadRecoveryKey()
	if err != nil {
		t.Fatalf("LoadRecoveryKey() error = %v", err)
	}

	if loaded != encryptedKey {
		t.Errorf("LoadRecoveryKey() = %q, want %q", loaded, encryptedKey)
	}
}

func TestSaveAndLoadRecoveryHash(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	hash := base64.StdEncoding.EncodeToString([]byte("test-hash-32-bytes-long-test!!"))

	// Save hash
	err := SaveRecoveryHash(hash)
	if err != nil {
		t.Fatalf("SaveRecoveryHash() error = %v", err)
	}

	// Load hash
	loaded, err := LoadRecoveryHash()
	if err != nil {
		t.Fatalf("LoadRecoveryHash() error = %v", err)
	}

	if loaded != hash {
		t.Errorf("LoadRecoveryHash() = %q, want %q", loaded, hash)
	}
}

func TestIsVersionCheckDisabled(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Should be false initially
	if IsVersionCheckDisabled() {
		t.Error("IsVersionCheckDisabled() should be false initially")
	}

	// Disable version check
	err := DisableVersionCheck()
	if err != nil {
		t.Fatalf("DisableVersionCheck() error = %v", err)
	}

	// Should be true after disabling
	if !IsVersionCheckDisabled() {
		t.Error("IsVersionCheckDisabled() should be true after disabling")
	}

	// Enable version check
	err = EnableVersionCheck()
	if err != nil {
		t.Fatalf("EnableVersionCheck() error = %v", err)
	}

	// Should be false after enabling
	if IsVersionCheckDisabled() {
		t.Error("IsVersionCheckDisabled() should be false after enabling")
	}
}

func TestLoadErrorMessages(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Should return nil when no config exists
	messages := LoadErrorMessages()
	if messages != nil {
		t.Error("LoadErrorMessages() should return nil when config doesn't exist")
	}
}

func TestLoadErrorTips(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Should return nil when no config exists
	tips := LoadErrorTips()
	if tips != nil {
		t.Error("LoadErrorTips() should return nil when config doesn't exist")
	}
}

func TestConfigPersistence(t *testing.T) {
	// Use temp home for testing
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Save salt
	salt := []byte("test-salt-32-bytes-long-test!")
	if err := SaveSalt(salt); err != nil {
		t.Fatalf("SaveSalt() error = %v", err)
	}

	// Save KDF version
	if err := SaveKDFVersion(2); err != nil {
		t.Fatalf("SaveKDFVersion() error = %v", err)
	}

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify values
	if string(cfg.Salt) != string(salt) {
		t.Error("Config salt does not match saved salt")
	}

	if cfg.KDFVersion != 2 {
		t.Errorf("Config KDFVersion = %d, want 2", cfg.KDFVersion)
	}

	// Verify keybindings were loaded (should be defaults)
	if cfg.Keybindings.Quit == "" {
		t.Error("Config keybindings not loaded")
	}
}
