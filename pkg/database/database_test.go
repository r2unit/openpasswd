package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/r2unit/openpasswd/pkg/models"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("New() returned nil database")
	}

	if db.path != dbPath {
		t.Errorf("New() path = %q, want %q", db.path, dbPath)
	}

	if db.nextID != 1 {
		t.Errorf("New() nextID = %d, want 1", db.nextID)
	}
}

func TestNewExistingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create initial database
	db1, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Add a password
	p := &models.Password{
		Name:     "Test",
		Username: "user",
		Password: "pass",
	}
	if err := db1.AddPassword(p); err != nil {
		t.Fatalf("AddPassword() error = %v", err)
	}
	db1.Close()

	// Open existing database
	db2, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() for existing db error = %v", err)
	}
	defer db2.Close()

	// Should load existing data
	passwords, err := db2.ListPasswords()
	if err != nil {
		t.Fatalf("ListPasswords() error = %v", err)
	}

	if len(passwords) != 1 {
		t.Errorf("ListPasswords() count = %d, want 1", len(passwords))
	}

	if db2.nextID != 2 {
		t.Errorf("New() nextID after load = %d, want 2", db2.nextID)
	}
}

func TestAddPassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	p := &models.Password{
		Name:     "GitHub",
		Username: "testuser",
		Password: "testpass",
		URL:      "https://github.com",
		Notes:    "Test notes",
	}

	err := db.AddPassword(p)
	if err != nil {
		t.Fatalf("AddPassword() error = %v", err)
	}

	// Check ID was assigned
	if p.ID == 0 {
		t.Error("AddPassword() did not assign ID")
	}

	// Check timestamps were set
	if p.CreatedAt.IsZero() {
		t.Error("AddPassword() did not set CreatedAt")
	}

	if p.UpdatedAt.IsZero() {
		t.Error("AddPassword() did not set UpdatedAt")
	}

	// Verify it's in the database
	retrieved, err := db.GetPassword(p.ID)
	if err != nil {
		t.Fatalf("GetPassword() error = %v", err)
	}

	if retrieved.Name != p.Name {
		t.Errorf("Retrieved name = %q, want %q", retrieved.Name, p.Name)
	}
}

func TestAddMultiplePasswords(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	passwords := []*models.Password{
		{Name: "Password 1", Username: "user1", Password: "pass1"},
		{Name: "Password 2", Username: "user2", Password: "pass2"},
		{Name: "Password 3", Username: "user3", Password: "pass3"},
	}

	for _, p := range passwords {
		if err := db.AddPassword(p); err != nil {
			t.Fatalf("AddPassword() error = %v", err)
		}
	}

	// IDs should be sequential
	for i, p := range passwords {
		expectedID := int64(i + 1)
		if p.ID != expectedID {
			t.Errorf("Password %d: ID = %d, want %d", i, p.ID, expectedID)
		}
	}

	// Next ID should be correct
	if db.nextID != 4 {
		t.Errorf("nextID = %d, want 4", db.nextID)
	}
}

func TestGetPassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	p := &models.Password{
		Name:     "Test",
		Username: "user",
		Password: "pass",
	}
	db.AddPassword(p)

	// Get existing password
	retrieved, err := db.GetPassword(p.ID)
	if err != nil {
		t.Fatalf("GetPassword() error = %v", err)
	}

	if retrieved.Name != p.Name {
		t.Errorf("GetPassword() name = %q, want %q", retrieved.Name, p.Name)
	}

	// Get non-existent password
	_, err = db.GetPassword(9999)
	if err == nil {
		t.Error("GetPassword() with invalid ID should return error")
	}
}

func TestListPasswords(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	// Empty list
	passwords, err := db.ListPasswords()
	if err != nil {
		t.Fatalf("ListPasswords() error = %v", err)
	}

	if len(passwords) != 0 {
		t.Errorf("ListPasswords() on empty db = %d passwords, want 0", len(passwords))
	}

	// Add some passwords
	for i := 0; i < 3; i++ {
		p := &models.Password{
			Name:     "Password " + string(rune('A'+i)),
			Username: "user",
			Password: "pass",
		}
		db.AddPassword(p)
	}

	// List should return all passwords
	passwords, err = db.ListPasswords()
	if err != nil {
		t.Fatalf("ListPasswords() error = %v", err)
	}

	if len(passwords) != 3 {
		t.Errorf("ListPasswords() count = %d, want 3", len(passwords))
	}
}

func TestUpdatePassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	p := &models.Password{
		Name:     "Original",
		Username: "user",
		Password: "pass",
	}
	db.AddPassword(p)

	// Update password
	time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt changes
	p.Name = "Updated"
	p.Username = "newuser"

	err := db.UpdatePassword(p)
	if err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	// Retrieve and verify
	retrieved, _ := db.GetPassword(p.ID)
	if retrieved.Name != "Updated" {
		t.Errorf("UpdatePassword() name = %q, want %q", retrieved.Name, "Updated")
	}

	if retrieved.Username != "newuser" {
		t.Errorf("UpdatePassword() username = %q, want %q", retrieved.Username, "newuser")
	}

	// UpdatedAt should change
	if !retrieved.UpdatedAt.After(retrieved.CreatedAt) {
		t.Error("UpdatePassword() should update UpdatedAt timestamp")
	}
}

func TestUpdatePasswordNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	p := &models.Password{
		ID:       9999,
		Name:     "Test",
		Username: "user",
		Password: "pass",
	}

	err := db.UpdatePassword(p)
	if err == nil {
		t.Error("UpdatePassword() with non-existent ID should return error")
	}
}

func TestDeletePassword(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	p := &models.Password{
		Name:     "Test",
		Username: "user",
		Password: "pass",
	}
	db.AddPassword(p)

	// Delete password
	err := db.DeletePassword(p.ID)
	if err != nil {
		t.Fatalf("DeletePassword() error = %v", err)
	}

	// Should not be retrievable
	_, err = db.GetPassword(p.ID)
	if err == nil {
		t.Error("GetPassword() should fail after deletion")
	}

	// List should be empty
	passwords, _ := db.ListPasswords()
	if len(passwords) != 0 {
		t.Errorf("ListPasswords() after delete = %d, want 0", len(passwords))
	}
}

func TestDeletePasswordNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	err := db.DeletePassword(9999)
	if err == nil {
		t.Error("DeletePassword() with non-existent ID should return error")
	}
}

func TestSearchPasswords(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	// Add test data
	passwords := []*models.Password{
		{Name: "GitHub", Username: "user1", Password: "pass1", URL: "https://github.com"},
		{Name: "GitLab", Username: "user2", Password: "pass2", URL: "https://gitlab.com"},
		{Name: "Email", Username: "test@gmail.com", Password: "pass3", URL: "https://gmail.com"},
		{Name: "Other", Username: "other", Password: "pass4", URL: "https://example.com"},
	}

	for _, p := range passwords {
		db.AddPassword(p)
	}

	tests := []struct {
		name     string
		query    string
		wantMin  int
		contains []string
	}{
		{"search by name", "git", 2, []string{"GitHub", "GitLab"}},
		{"search by username", "user1", 1, []string{"GitHub"}},
		{"search by URL", "gmail", 1, []string{"Email"}},
		{"case insensitive", "GITHUB", 1, []string{"GitHub"}},
		{"no results", "nonexistent", 0, nil},
		{"partial match", "git", 2, []string{"GitHub", "GitLab"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.SearchPasswords(tt.query)
			if err != nil {
				t.Fatalf("SearchPasswords() error = %v", err)
			}

			if len(results) < tt.wantMin {
				t.Errorf("SearchPasswords() found %d results, want at least %d", len(results), tt.wantMin)
			}

			// Check that expected items are in results
			for _, expectedName := range tt.contains {
				found := false
				for _, result := range results {
					if result.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("SearchPasswords() missing expected result: %q", expectedName)
				}
			}
		})
	}
}

func TestSearchPasswordsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	// Search empty database
	results, err := db.SearchPasswords("test")
	if err != nil {
		t.Fatalf("SearchPasswords() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("SearchPasswords() on empty db = %d results, want 0", len(results))
	}
}

func TestDatabasePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database and add password
	db1, _ := New(dbPath)
	p := &models.Password{
		Name:     "Persistent",
		Username: "user",
		Password: "pass",
	}
	db1.AddPassword(p)
	db1.Close()

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file was not created")
	}

	// Reopen database
	db2, _ := New(dbPath)
	defer db2.Close()

	// Data should be loaded
	passwords, _ := db2.ListPasswords()
	if len(passwords) != 1 {
		t.Fatalf("After reload: got %d passwords, want 1", len(passwords))
	}

	if passwords[0].Name != "Persistent" {
		t.Errorf("After reload: name = %q, want %q", passwords[0].Name, "Persistent")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	// Add initial password
	p := &models.Password{
		Name:     "Test",
		Username: "user",
		Password: "pass",
	}
	db.AddPassword(p)

	// Concurrent reads should work
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := db.GetPassword(p.ID)
			if err != nil {
				t.Errorf("Concurrent GetPassword() error = %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestPasswordFields(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, _ := New(dbPath)
	defer db.Close()

	p := &models.Password{
		Name:     "Test",
		Username: "user",
		Password: "pass",
		Fields: map[string]string{
			"security_question": "What is your pet's name?",
			"answer":            "Fluffy",
		},
	}

	err := db.AddPassword(p)
	if err != nil {
		t.Fatalf("AddPassword() error = %v", err)
	}

	// Retrieve and check fields
	retrieved, _ := db.GetPassword(p.ID)
	if len(retrieved.Fields) != 2 {
		t.Errorf("Retrieved fields count = %d, want 2", len(retrieved.Fields))
	}

	if retrieved.Fields["security_question"] != "What is your pet's name?" {
		t.Error("Fields were not preserved correctly")
	}
}
