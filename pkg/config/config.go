package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/r2unit/openpasswd/pkg/toml"
)

type Keybindings struct {
	Quit    string `toml:"quit"`
	QuitAlt string `toml:"quit_alt"`
	Back    string `toml:"back"`
	Up      string `toml:"up"`
	UpAlt   string `toml:"up_alt"`
	Down    string `toml:"down"`
	DownAlt string `toml:"down_alt"`
	Select  string `toml:"select"`
}

type Config struct {
	DatabasePath string
	Salt         []byte
	Keybindings  Keybindings
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "openpasswd")
	return configDir, nil
}

func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return configDir, nil
}

func DefaultKeybindings() Keybindings {
	return Keybindings{
		Quit:    ":q",
		QuitAlt: "ctrl+c",
		Back:    "esc",
		Up:      "up",
		UpAlt:   "k",
		Down:    "down",
		DownAlt: "j",
		Select:  "enter",
	}
}

func LoadKeybindings() (Keybindings, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return DefaultKeybindings(), nil
	}

	configPath := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultKeybindings(), nil
	}

	type ConfigFile struct {
		Keybindings Keybindings `toml:"keybindings"`
	}

	var cfg ConfigFile
	_, err = toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return DefaultKeybindings(), nil
	}

	kb := DefaultKeybindings()
	if cfg.Keybindings.Quit != "" {
		kb.Quit = cfg.Keybindings.Quit
	}
	if cfg.Keybindings.QuitAlt != "" {
		kb.QuitAlt = cfg.Keybindings.QuitAlt
	}
	if cfg.Keybindings.Back != "" {
		kb.Back = cfg.Keybindings.Back
	}
	if cfg.Keybindings.Up != "" {
		kb.Up = cfg.Keybindings.Up
	}
	if cfg.Keybindings.UpAlt != "" {
		kb.UpAlt = cfg.Keybindings.UpAlt
	}
	if cfg.Keybindings.Down != "" {
		kb.Down = cfg.Keybindings.Down
	}
	if cfg.Keybindings.DownAlt != "" {
		kb.DownAlt = cfg.Keybindings.DownAlt
	}
	if cfg.Keybindings.Select != "" {
		kb.Select = cfg.Keybindings.Select
	}

	return kb, nil
}

func LoadConfig() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	saltPath := filepath.Join(configDir, "salt")
	saltData, err := os.ReadFile(saltPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration not initialized, please run 'openpass init'")
		}
		return nil, err
	}

	salt, err := base64.StdEncoding.DecodeString(string(saltData))
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(configDir, "passwords.db")

	keybindings, err := LoadKeybindings()
	if err != nil {
		keybindings = DefaultKeybindings()
	}

	return &Config{
		DatabasePath: dbPath,
		Salt:         salt,
		Keybindings:  keybindings,
	}, nil
}

func SaveSalt(salt []byte) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	saltPath := filepath.Join(configDir, "salt")
	encoded := base64.StdEncoding.EncodeToString(salt)
	return os.WriteFile(saltPath, []byte(encoded), 0600)
}

func HasPassphrase() bool {
	configDir, err := GetConfigDir()
	if err != nil {
		return false
	}

	passphrasePath := filepath.Join(configDir, "passphrase")
	_, err = os.Stat(passphrasePath)
	return err == nil
}

func SavePassphrase(passphrase string) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	passphrasePath := filepath.Join(configDir, "passphrase")
	return os.WriteFile(passphrasePath, []byte(passphrase), 0600)
}

func LoadPassphrase() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	passphrasePath := filepath.Join(configDir, "passphrase")
	data, err := os.ReadFile(passphrasePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func RemovePassphrase() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	passphrasePath := filepath.Join(configDir, "passphrase")
	err = os.Remove(passphrasePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func HasTOTP() bool {
	configDir, err := GetConfigDir()
	if err != nil {
		return false
	}

	totpPath := filepath.Join(configDir, "totp_secret")
	_, err = os.Stat(totpPath)
	return err == nil
}

func SaveTOTPSecret(secret string) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	totpPath := filepath.Join(configDir, "totp_secret")
	encoded := base64.StdEncoding.EncodeToString([]byte(secret))
	return os.WriteFile(totpPath, []byte(encoded), 0600)
}

func LoadTOTPSecret() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	totpPath := filepath.Join(configDir, "totp_secret")
	data, err := os.ReadFile(totpPath)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func RemoveTOTP() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	totpPath := filepath.Join(configDir, "totp_secret")
	err = os.Remove(totpPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func HasYubiKey() bool {
	configDir, err := GetConfigDir()
	if err != nil {
		return false
	}

	ykPath := filepath.Join(configDir, "yubikey_challenge")
	_, err = os.Stat(ykPath)
	return err == nil
}

func SaveYubiKeyChallenge(challenge string) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	ykPath := filepath.Join(configDir, "yubikey_challenge")
	return os.WriteFile(ykPath, []byte(challenge), 0600)
}

func LoadYubiKeyChallenge() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	ykPath := filepath.Join(configDir, "yubikey_challenge")
	data, err := os.ReadFile(ykPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func RemoveYubiKey() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	ykPath := filepath.Join(configDir, "yubikey_challenge")
	err = os.Remove(ykPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

