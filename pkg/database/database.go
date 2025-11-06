package database

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/r2unit/openpasswd/pkg/models"
)

type DB struct {
	path      string
	passwords map[int64]*models.Password
	nextID    int64
	mu        sync.RWMutex
}

func New(dbPath string) (*DB, error) {
	db := &DB{
		path:      dbPath,
		passwords: make(map[int64]*models.Password),
		nextID:    1,
	}

	if err := db.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return db, nil
}

func (db *DB) load() error {
	data, err := os.ReadFile(db.path)
	if err != nil {
		return err
	}

	var store struct {
		NextID    int64                      `json:"next_id"`
		Passwords map[int64]*models.Password `json:"passwords"`
	}

	if err := json.Unmarshal(data, &store); err != nil {
		return err
	}

	db.passwords = store.Passwords
	db.nextID = store.NextID

	return nil
}

func (db *DB) save() error {
	store := struct {
		NextID    int64                      `json:"next_id"`
		Passwords map[int64]*models.Password `json:"passwords"`
	}{
		NextID:    db.nextID,
		Passwords: db.passwords,
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, data, 0600)
}

func (db *DB) Close() error {
	return nil
}

func (db *DB) AddPassword(p *models.Password) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	now := time.Now()
	p.ID = db.nextID
	p.CreatedAt = now
	p.UpdatedAt = now

	db.passwords[p.ID] = p
	db.nextID++

	return db.save()
}

func (db *DB) GetPassword(id int64) (*models.Password, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	p, ok := db.passwords[id]
	if !ok {
		return nil, errors.New("password not found")
	}

	return p, nil
}

func (db *DB) ListPasswords() ([]*models.Password, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	passwords := make([]*models.Password, 0, len(db.passwords))
	for _, p := range db.passwords {
		passwords = append(passwords, p)
	}

	return passwords, nil
}

func (db *DB) UpdatePassword(p *models.Password) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.passwords[p.ID]; !ok {
		return errors.New("password not found")
	}

	p.UpdatedAt = time.Now()
	db.passwords[p.ID] = p

	return db.save()
}

func (db *DB) DeletePassword(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.passwords[id]; !ok {
		return errors.New("password not found")
	}

	delete(db.passwords, id)
	return db.save()
}

func (db *DB) SearchPasswords(search string) ([]*models.Password, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	search = strings.ToLower(search)
	var passwords []*models.Password

	for _, p := range db.passwords {
		if strings.Contains(strings.ToLower(p.Name), search) ||
			strings.Contains(strings.ToLower(p.Username), search) ||
			strings.Contains(strings.ToLower(p.URL), search) {
			passwords = append(passwords, p)
		}
	}

	return passwords, nil
}
