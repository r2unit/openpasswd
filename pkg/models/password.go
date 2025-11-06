package models

import "time"

type PasswordType string

const (
	TypeLogin    PasswordType = "login"
	TypeCard     PasswordType = "card"
	TypeNote     PasswordType = "note"
	TypeIdentity PasswordType = "identity"
	TypeOther    PasswordType = "other"
	TypePassword PasswordType = "password"
)

type Password struct {
	ID        int64
	Type      PasswordType
	Name      string
	Username  string
	Password  string
	URL       string
	Notes     string
	Fields    map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}
