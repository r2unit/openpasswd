package database

import (
	"github.com/r2unit/openpasswd/pkg/crypto"
)

// SaveIntegrityCheck saves an HMAC for the database file
func (db *DB) SaveIntegrityCheck(encryptor *crypto.Encryptor) error {
	return crypto.SaveDatabaseHMAC(db.path, encryptor)
}

// VerifyIntegrityCheck verifies the HMAC of the database file
func (db *DB) VerifyIntegrityCheck(encryptor *crypto.Encryptor) error {
	return crypto.VerifyDatabaseHMAC(db.path, encryptor)
}
