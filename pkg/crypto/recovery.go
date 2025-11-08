package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

// BIP39-style word list (simplified, first 256 words for demonstration)
// In production, use full 2048-word BIP39 list
var recoveryWords = []string{
	"abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract",
	"absurd", "abuse", "access", "accident", "account", "accuse", "achieve", "acid",
	"acoustic", "acquire", "across", "act", "action", "actor", "actress", "actual",
	"adapt", "add", "addict", "address", "adjust", "admit", "adult", "advance",
	"advice", "aerobic", "afford", "afraid", "again", "age", "agent", "agree",
	"ahead", "aim", "air", "airport", "aisle", "alarm", "album", "alcohol",
	"alert", "alien", "all", "alley", "allow", "almost", "alone", "alpha",
	"already", "also", "alter", "always", "amateur", "amazing", "among", "amount",
	"amused", "analyst", "anchor", "ancient", "anger", "angle", "angry", "animal",
	"ankle", "announce", "annual", "another", "answer", "antenna", "antique", "anxiety",
	"any", "apart", "apology", "appear", "apple", "approve", "april", "arch",
	"arctic", "area", "arena", "argue", "arm", "armed", "armor", "army",
	"around", "arrange", "arrest", "arrive", "arrow", "art", "artefact", "artist",
	"artwork", "ask", "aspect", "assault", "asset", "assist", "assume", "asthma",
	"athlete", "atom", "attack", "attend", "attitude", "attract", "auction", "audit",
	"august", "aunt", "author", "auto", "autumn", "average", "avocado", "avoid",
	"awake", "aware", "away", "awesome", "awful", "awkward", "axis", "baby",
	"bachelor", "bacon", "badge", "bag", "balance", "balcony", "ball", "bamboo",
	"banana", "banner", "bar", "barely", "bargain", "barrel", "base", "basic",
	"basket", "battle", "beach", "bean", "beauty", "because", "become", "beef",
	"before", "begin", "behave", "behind", "believe", "below", "belt", "bench",
	"benefit", "best", "betray", "better", "between", "beyond", "bicycle", "bid",
	"bike", "bind", "biology", "bird", "birth", "bitter", "black", "blade",
	"blame", "blanket", "blast", "bleak", "bless", "blind", "blood", "blossom",
	"blouse", "blue", "blur", "blush", "board", "boat", "body", "boil",
	"bomb", "bone", "bonus", "book", "boost", "border", "boring", "borrow",
	"boss", "bottom", "bounce", "box", "boy", "bracket", "brain", "brand",
	"brass", "brave", "bread", "breeze", "brick", "bridge", "brief", "bright",
	"bring", "brisk", "broccoli", "broken", "bronze", "broom", "brother", "brown",
	"brush", "bubble", "buddy", "budget", "buffalo", "build", "bulb", "bulk",
	"bullet", "bundle", "bunker", "burden", "burger", "burst", "bus", "business",
	"busy", "butter", "buyer", "buzz", "cabbage", "cabin", "cable", "cactus",
}

// GenerateRecoveryKey generates a cryptographically secure recovery key
// Returns a 24-word mnemonic phrase
func GenerateRecoveryKey() (string, error) {
	// Generate 32 bytes of entropy (256 bits)
	entropy := make([]byte, 32)
	if _, err := rand.Read(entropy); err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Convert entropy to word indices
	words := make([]string, 24)
	wordCount := len(recoveryWords)

	for i := 0; i < 24; i++ {
		// Use modulo to map byte to word index
		index := int(entropy[i]) % wordCount
		words[i] = recoveryWords[index]
	}

	return strings.Join(words, "-"), nil
}

// RecoveryKeyToSeed converts a recovery key mnemonic to a seed
func RecoveryKeyToSeed(recoveryKey string) ([]byte, error) {
	words := strings.Split(recoveryKey, "-")
	if len(words) != 24 {
		return nil, fmt.Errorf("invalid recovery key: expected 24 words, got %d", len(words))
	}

	// Convert words back to entropy
	entropy := make([]byte, 32)
	for i, word := range words {
		// Find word index
		index := -1
		for j, w := range recoveryWords {
			if w == word {
				index = j
				break
			}
		}
		if index == -1 {
			return nil, fmt.Errorf("invalid word in recovery key: %s", word)
		}
		entropy[i] = byte(index)
	}

	// Derive a key from the entropy using Blake2b
	seed := Blake2bSum256(entropy)

	return seed, nil
}

// DerivePassphraseFromRecovery derives an encryption key from a recovery key
// This allows using the recovery key to decrypt the database
func DerivePassphraseFromRecovery(recoveryKey string, salt []byte) ([]byte, error) {
	seed, err := RecoveryKeyToSeed(recoveryKey)
	if err != nil {
		return nil, err
	}

	// Use Argon2id to derive a strong key from the recovery seed
	params := DefaultArgon2Params()
	key := Argon2idKey(seed, salt, params)

	return key, nil
}

// EncryptRecoveryKey encrypts the recovery key with the user's passphrase
// This allows storing the recovery key encrypted in the config
func EncryptRecoveryKey(recoveryKey, passphrase string, salt []byte) (string, error) {
	encryptor := NewEncryptor(passphrase, salt)
	encrypted, err := encryptor.Encrypt(recoveryKey)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

// DecryptRecoveryKey decrypts a stored recovery key
func DecryptRecoveryKey(encryptedKey, passphrase string, salt []byte) (string, error) {
	encryptor := NewEncryptor(passphrase, salt)
	decrypted, err := encryptor.Decrypt(encryptedKey)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

// FormatRecoveryKey formats a recovery key for display (groups of 4 words)
func FormatRecoveryKey(key string) string {
	words := strings.Split(key, "-")
	if len(words) != 24 {
		return key
	}

	var formatted strings.Builder
	for i := 0; i < 6; i++ {
		start := i * 4
		end := start + 4
		if end > len(words) {
			end = len(words)
		}
		formatted.WriteString(fmt.Sprintf("%2d. %s\n", i*4+1, strings.Join(words[start:end], "-")))
	}

	return formatted.String()
}

// GenerateRecoveryHash generates a hash of the recovery key for verification
// This can be stored to verify if a recovery key is correct without storing the key itself
func GenerateRecoveryHash(recoveryKey string) []byte {
	seed, _ := RecoveryKeyToSeed(recoveryKey)
	return Blake2bSum256(seed)
}

// VerifyRecoveryKey verifies if a recovery key matches the stored hash
func VerifyRecoveryKey(recoveryKey string, hash []byte) bool {
	calculatedHash := GenerateRecoveryHash(recoveryKey)
	if len(calculatedHash) != len(hash) {
		return false
	}
	for i := range hash {
		if hash[i] != calculatedHash[i] {
			return false
		}
	}
	return true
}

// EncodeRecoveryHash encodes a recovery key hash for storage
func EncodeRecoveryHash(hash []byte) string {
	return base64.StdEncoding.EncodeToString(hash)
}

// DecodeRecoveryHash decodes a stored recovery key hash
func DecodeRecoveryHash(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// GenerateRecoveryKeyWithChecksum generates a recovery key with a checksum word
func GenerateRecoveryKeyWithChecksum() (string, error) {
	key, err := GenerateRecoveryKey()
	if err != nil {
		return "", err
	}

	// Calculate checksum (first word of hash)
	hash := GenerateRecoveryHash(key)
	checksumIndex := binary.BigEndian.Uint16(hash[:2]) % uint16(len(recoveryWords))
	checksumWord := recoveryWords[checksumIndex]

	return key + "-" + checksumWord, nil
}

// VerifyRecoveryKeyChecksum verifies the checksum of a recovery key
func VerifyRecoveryKeyChecksum(keyWithChecksum string) (bool, string) {
	parts := strings.Split(keyWithChecksum, "-")
	if len(parts) != 25 {
		return false, ""
	}

	key := strings.Join(parts[:24], "-")
	providedChecksum := parts[24]

	hash := GenerateRecoveryHash(key)
	checksumIndex := binary.BigEndian.Uint16(hash[:2]) % uint16(len(recoveryWords))
	expectedChecksum := recoveryWords[checksumIndex]

	return providedChecksum == expectedChecksum, key
}
