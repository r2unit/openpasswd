package server

// TODO: Remote Server Package - Password sync server implementation
// This package provides HTTP server functionality for syncing passwords across devices
// Currently disabled - future feature for multi-device synchronization
//
// Planned features:
// - End-to-end encryption with client-side keys
// - RESTful API for password CRUD operations
// - Session management with token-based authentication
// - Conflict resolution for simultaneous edits
// - WebSocket support for real-time sync
//
// Security considerations:
// - Zero-knowledge architecture (server never sees plaintext)
// - TLS/SSL required for all connections
// - Rate limiting to prevent brute force attacks
// - Audit logging for all operations

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/r2unit/openpasswd/pkg/auth"
	"github.com/r2unit/openpasswd/pkg/crypto"
	"github.com/r2unit/openpasswd/pkg/database"
	"github.com/r2unit/openpasswd/pkg/models"
)

type Server struct {
	db        *database.DB
	salt      []byte
	sessions  map[string]*auth.Session
	mu        sync.RWMutex
	masterKey string
}

func New(db *database.DB, salt []byte, masterKey string) *Server {
	return &Server{
		db:        db,
		salt:      salt,
		sessions:  make(map[string]*auth.Session),
		masterKey: masterKey,
	}
}

func (s *Server) Start(port string) error {
	http.HandleFunc("/api/auth/login", s.handleLogin)
	http.HandleFunc("/api/auth/logout", s.handleLogout)
	http.HandleFunc("/api/passwords", s.handlePasswords)
	http.HandleFunc("/api/passwords/", s.handlePassword)
	http.HandleFunc("/api/passwords/search", s.handleSearch)
	http.HandleFunc("/api/health", s.handleHealth)

	addr := ":" + port
	fmt.Printf("Server starting on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) authenticate(r *http.Request) (*auth.Session, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	if len(tokenHeader) < 7 || tokenHeader[:7] != "Bearer " {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token := tokenHeader[7:]
	tokenHash := auth.HashToken(token)

	s.mu.RLock()
	session, ok := s.sessions[tokenHash]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, tokenHash)
		s.mu.Unlock()
		return nil, fmt.Errorf("token expired")
	}

	return session, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Passphrase string `json:"passphrase"`
		MasterKey  string `json:"master_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.MasterKey != s.masterKey {
		http.Error(w, "Invalid master key", http.StatusUnauthorized)
		return
	}

	encryptor := crypto.NewEncryptor(req.Passphrase, s.salt)
	testPassword := &models.Password{
		ID:       -1,
		Name:     "test",
		Username: "test",
		Password: "test",
	}

	encrypted, err := encryptor.Encrypt(testPassword.Password)
	if err != nil {
		http.Error(w, "Encryption test failed", http.StatusInternalServerError)
		return
	}

	if _, err := encryptor.Decrypt(encrypted); err != nil {
		http.Error(w, "Invalid passphrase", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	tokenHash := auth.HashToken(token)

	s.mu.Lock()
	s.sessions[tokenHash] = &auth.Session{
		Token:      token,
		Passphrase: req.Passphrase,
		ExpiresAt:  expiresAt,
	}
	s.mu.Unlock()

	log.Printf("New session created, expires at %s", expiresAt.Format(time.RFC3339))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tokenHash := auth.HashToken(session.Token)
	s.mu.Lock()
	delete(s.sessions, tokenHash)
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "logged out"})
}

func (s *Server) handlePasswords(w http.ResponseWriter, r *http.Request) {
	session, err := s.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	encryptor := crypto.NewEncryptor(session.Passphrase, s.salt)

	switch r.Method {
	case http.MethodGet:
		passwords, err := s.db.ListPasswords()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, p := range passwords {
			decrypted, err := encryptor.Decrypt(p.Password)
			if err != nil {
				http.Error(w, "Failed to decrypt password", http.StatusInternalServerError)
				return
			}
			p.Password = decrypted
		}

		json.NewEncoder(w).Encode(passwords)

	case http.MethodPost:
		var p models.Password
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		encrypted, err := encryptor.Encrypt(p.Password)
		if err != nil {
			http.Error(w, "Failed to encrypt password", http.StatusInternalServerError)
			return
		}
		p.Password = encrypted

		if err := s.db.AddPassword(&p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePassword(w http.ResponseWriter, r *http.Request) {
	session, err := s.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	encryptor := crypto.NewEncryptor(session.Passphrase, s.salt)

	var id int64
	fmt.Sscanf(r.URL.Path, "/api/passwords/%d", &id)

	switch r.Method {
	case http.MethodGet:
		p, err := s.db.GetPassword(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		decrypted, err := encryptor.Decrypt(p.Password)
		if err != nil {
			http.Error(w, "Failed to decrypt password", http.StatusInternalServerError)
			return
		}
		p.Password = decrypted

		json.NewEncoder(w).Encode(p)

	case http.MethodPut:
		var p models.Password
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		p.ID = id

		encrypted, err := encryptor.Encrypt(p.Password)
		if err != nil {
			http.Error(w, "Failed to encrypt password", http.StatusInternalServerError)
			return
		}
		p.Password = encrypted

		if err := s.db.UpdatePassword(&p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(p)

	case http.MethodDelete:
		if err := s.db.DeletePassword(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	session, err := s.authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	encryptor := crypto.NewEncryptor(session.Passphrase, s.salt)

	passwords, err := s.db.SearchPasswords(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, p := range passwords {
		decrypted, err := encryptor.Decrypt(p.Password)
		if err != nil {
			http.Error(w, "Failed to decrypt password", http.StatusInternalServerError)
			return
		}
		p.Password = decrypted
	}

	json.NewEncoder(w).Encode(passwords)
}
