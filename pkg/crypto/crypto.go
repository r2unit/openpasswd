package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"hash"
	"io"
)

const (
	saltSize = 32
	keySize  = 32
	// Legacy iteration count (deprecated)
	iterationsLegacy = 100000
	// Current iteration count (OWASP recommended)
	iterationsCurrent = 600000
)

type Encryptor struct {
	key []byte
}

// GetKey returns the encryption key (for HMAC derivation)
func (e *Encryptor) GetKey() []byte {
	return e.key
}

func pbkdf2Key(password, salt []byte, iterations, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)

	for block := 1; block <= numBlocks; block++ {
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		u := prf.Sum(nil)
		out := make([]byte, len(u))
		copy(out, u)

		for i := 2; i <= iterations; i++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(nil)
			for j := range out {
				out[j] ^= u[j]
			}
		}

		dk = append(dk, out...)
	}

	return dk[:keyLen]
}

// NewEncryptor creates an encryptor using the current KDF version (600k iterations)
func NewEncryptor(passphrase string, salt []byte) *Encryptor {
	key := pbkdf2Key([]byte(passphrase), salt, iterationsCurrent, keySize, sha256.New)
	return &Encryptor{key: key}
}

// NewEncryptorWithVersion creates an encryptor with a specific KDF version
func NewEncryptorWithVersion(passphrase string, salt []byte, version int) *Encryptor {
	params := GetKDFParams(version)
	key := pbkdf2Key([]byte(passphrase), salt, params.Iterations, keySize, sha256.New)
	return &Encryptor{key: key}
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
