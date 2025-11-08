package crypto

// KDF version constants for backward compatibility
const (
	KDFVersionPBKDF2_100k = 1 // Legacy (100,000 iterations)
	KDFVersionPBKDF2_600k = 2 // Current (600,000 iterations)
	KDFVersionArgon2id    = 3 // Future (memory-hard KDF)

	// CurrentKDFVersion is the default version for new encryptions
	CurrentKDFVersion = KDFVersionPBKDF2_600k
)

// KDFParams holds parameters for key derivation
type KDFParams struct {
	Version    int
	Iterations int
	// Future: Memory, Threads for Argon2
}

// GetKDFParams returns the parameters for a given KDF version
func GetKDFParams(version int) KDFParams {
	switch version {
	case KDFVersionPBKDF2_100k:
		return KDFParams{
			Version:    KDFVersionPBKDF2_100k,
			Iterations: 100000,
		}
	case KDFVersionPBKDF2_600k:
		return KDFParams{
			Version:    KDFVersionPBKDF2_600k,
			Iterations: 600000,
		}
	case KDFVersionArgon2id:
		// Future implementation
		return KDFParams{
			Version:    KDFVersionArgon2id,
			Iterations: 0, // Argon2 doesn't use iterations the same way
		}
	default:
		// Default to current version
		return GetKDFParams(CurrentKDFVersion)
	}
}
