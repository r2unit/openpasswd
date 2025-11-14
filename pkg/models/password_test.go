package models

import (
	"testing"
	"time"
)

func TestPasswordType(t *testing.T) {
	tests := []struct {
		name     string
		pwdType  PasswordType
		expected string
	}{
		{"login type", TypeLogin, "login"},
		{"card type", TypeCard, "card"},
		{"note type", TypeNote, "note"},
		{"identity type", TypeIdentity, "identity"},
		{"other type", TypeOther, "other"},
		{"password type", TypePassword, "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.pwdType) != tt.expected {
				t.Errorf("PasswordType %q = %q, want %q", tt.name, tt.pwdType, tt.expected)
			}
		})
	}
}

func TestPasswordStruct(t *testing.T) {
	now := time.Now()

	p := Password{
		ID:        1,
		Type:      TypeLogin,
		Name:      "Test Password",
		Username:  "testuser",
		Password:  "testpass",
		URL:       "https://example.com",
		Notes:     "Test notes",
		Fields:    map[string]string{"key": "value"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test all fields are set correctly
	if p.ID != 1 {
		t.Errorf("Password.ID = %d, want 1", p.ID)
	}

	if p.Type != TypeLogin {
		t.Errorf("Password.Type = %q, want %q", p.Type, TypeLogin)
	}

	if p.Name != "Test Password" {
		t.Errorf("Password.Name = %q, want %q", p.Name, "Test Password")
	}

	if p.Username != "testuser" {
		t.Errorf("Password.Username = %q, want %q", p.Username, "testuser")
	}

	if p.Password != "testpass" {
		t.Errorf("Password.Password = %q, want %q", p.Password, "testpass")
	}

	if p.URL != "https://example.com" {
		t.Errorf("Password.URL = %q, want %q", p.URL, "https://example.com")
	}

	if p.Notes != "Test notes" {
		t.Errorf("Password.Notes = %q, want %q", p.Notes, "Test notes")
	}

	if len(p.Fields) != 1 {
		t.Errorf("Password.Fields length = %d, want 1", len(p.Fields))
	}

	if p.Fields["key"] != "value" {
		t.Errorf("Password.Fields[key] = %q, want %q", p.Fields["key"], "value")
	}

	if p.CreatedAt.IsZero() {
		t.Error("Password.CreatedAt should not be zero")
	}

	if p.UpdatedAt.IsZero() {
		t.Error("Password.UpdatedAt should not be zero")
	}
}

func TestPasswordEmpty(t *testing.T) {
	p := Password{}

	// Check zero values
	if p.ID != 0 {
		t.Errorf("Empty Password.ID = %d, want 0", p.ID)
	}

	if p.Type != "" {
		t.Errorf("Empty Password.Type = %q, want empty", p.Type)
	}

	if p.Name != "" {
		t.Errorf("Empty Password.Name = %q, want empty", p.Name)
	}

	if p.CreatedAt.IsZero() == false {
		t.Error("Empty Password.CreatedAt should be zero")
	}
}

func TestPasswordWithFields(t *testing.T) {
	p := Password{
		Name:   "Test",
		Fields: make(map[string]string),
	}

	// Add fields
	p.Fields["email"] = "test@example.com"
	p.Fields["phone"] = "123-456-7890"
	p.Fields["security_question"] = "What is your pet's name?"

	// Verify fields
	if len(p.Fields) != 3 {
		t.Errorf("Password.Fields length = %d, want 3", len(p.Fields))
	}

	if p.Fields["email"] != "test@example.com" {
		t.Error("Password.Fields[email] incorrect")
	}

	if p.Fields["phone"] != "123-456-7890" {
		t.Error("Password.Fields[phone] incorrect")
	}
}

func TestPasswordTypeString(t *testing.T) {
	// Test that PasswordType can be converted to string
	pwdType := TypeLogin
	str := string(pwdType)

	if str != "login" {
		t.Errorf("string(TypeLogin) = %q, want %q", str, "login")
	}
}

func TestPasswordNilFields(t *testing.T) {
	// Test password with nil Fields map
	p := Password{
		Name:   "Test",
		Fields: nil,
	}

	// Should not panic when accessing
	if p.Fields != nil {
		t.Error("Password.Fields should be nil")
	}

	// Initialize map before use
	if p.Fields == nil {
		p.Fields = make(map[string]string)
	}

	p.Fields["test"] = "value"

	if p.Fields["test"] != "value" {
		t.Error("Failed to set field after initializing map")
	}
}
