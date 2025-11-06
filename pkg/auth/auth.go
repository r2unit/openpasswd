package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Token struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	ServerURL string    `json:"server_url"`
}

type Session struct {
	Token      string
	Passphrase string
	ExpiresAt  time.Time
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func SaveClientToken(serverURL, token string, expiresAt time.Time) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config", "passwd")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	tokenData := Token{
		Value:     token,
		ExpiresAt: expiresAt,
		ServerURL: serverURL,
	}

	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(configDir, "token.json")
	return os.WriteFile(tokenPath, data, 0600)
}

func LoadClientToken() (*Token, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	tokenPath := filepath.Join(home, ".config", "passwd", "token.json")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	return &token, nil
}

func DeleteClientToken() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(home, ".config", "passwd", "token.json")
	return os.Remove(tokenPath)
}
