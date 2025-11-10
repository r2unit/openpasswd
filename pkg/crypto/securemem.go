package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"sync"
)

// SecureString stores a string encrypted in memory
// This prevents passwords from being stored in plaintext in RAM
type SecureString struct {
	ciphertext []byte
	nonce      []byte
	mu         sync.RWMutex
}

// Memory encryption key (generated once per process)
var (
	memoryKey     []byte
	memoryKeyOnce sync.Once
)

// getMemoryKey returns the process-wide memory encryption key
func getMemoryKey() []byte {
	memoryKeyOnce.Do(func() {
		memoryKey = make([]byte, 32)
		if _, err := rand.Read(memoryKey); err != nil {
			panic("failed to generate memory encryption key: " + err.Error())
		}
	})
	return memoryKey
}

// NewSecureString creates a new secure string from plaintext
func NewSecureString(plaintext string) (*SecureString, error) {
	if plaintext == "" {
		return &SecureString{}, nil
	}

	key := getMemoryKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return &SecureString{
		ciphertext: ciphertext,
		nonce:      nonce,
	}, nil
}

// Get decrypts and returns the plaintext value
func (s *SecureString) Get() (string, error) {
	if s == nil || len(s.ciphertext) == 0 {
		return "", nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := getMemoryKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, s.nonce, s.ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Set encrypts and stores a new value
func (s *SecureString) Set(plaintext string) error {
	if plaintext == "" {
		s.mu.Lock()
		s.ciphertext = nil
		s.nonce = nil
		s.mu.Unlock()
		return nil
	}

	key := getMemoryKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	s.mu.Lock()
	s.ciphertext = ciphertext
	s.nonce = nonce
	s.mu.Unlock()

	return nil
}

// Wipe securely wipes the secure string from memory
func (s *SecureString) Wipe() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Overwrite with zeros
	for i := range s.ciphertext {
		s.ciphertext[i] = 0
	}
	for i := range s.nonce {
		s.nonce[i] = 0
	}

	s.ciphertext = nil
	s.nonce = nil
}

// IsEmpty returns true if the secure string is empty
func (s *SecureString) IsEmpty() bool {
	if s == nil {
		return true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.ciphertext) == 0
}

// MarshalJSON implements json.Marshaler for SecureString
// This allows passwords to be stored encrypted in the database
func (s *SecureString) MarshalJSON() ([]byte, error) {
	if s == nil || len(s.ciphertext) == 0 {
		return []byte(`""`), nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Combine nonce + ciphertext for storage
	combined := append(s.nonce, s.ciphertext...)
	encoded := base64.StdEncoding.EncodeToString(combined)

	return []byte(`"` + encoded + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for SecureString
func (s *SecureString) UnmarshalJSON(data []byte) error {
	// Remove quotes
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return nil
	}
	data = data[1 : len(data)-1]

	if len(data) == 0 {
		return nil
	}

	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	// Split nonce and ciphertext
	gcmNonceSize := 12 // Standard GCM nonce size
	if len(combined) < gcmNonceSize {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nonce = combined[:gcmNonceSize]
	s.ciphertext = combined[gcmNonceSize:]

	return nil
}

// WipeMemory overwrites a byte slice with random data then zeros
func WipeMemory(data []byte) {
	if len(data) == 0 {
		return
	}

	// Overwrite with random data
	_, _ = rand.Read(data) // Ignore error as this is best-effort wiping

	// Overwrite with zeros
	for i := range data {
		data[i] = 0
	}
}

// SecureHash computes a secure hash of sensitive data and wipes the input
func SecureHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	WipeMemory(data)
	return hash[:]
}
