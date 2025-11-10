package client

// This package provides HTTP client functionality for syncing with OpenPasswd server
//
// Planned features:
// - RESTful API client with automatic retry logic
// - Secure token-based authentication
// - Background sync with conflict detection
// - Offline mode with queue for pending changes
// - Delta sync to minimize bandwidth usage
//
// Usage example (future):
//   client := client.New("https://sync.openpasswd.com", token)
//   passwords, err := client.ListPasswords()
//   err = client.SyncPassword(password)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/r2unit/openpasswd/pkg/auth"
	"github.com/r2unit/openpasswd/pkg/models"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

func Login(serverURL, passphrase, masterKey string) (string, time.Time, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	reqBody := map[string]string{
		"passphrase": passphrase,
		"master_key": masterKey,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", time.Time{}, err
	}

	resp, err := client.Post(serverURL+"/api/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, fmt.Errorf("login failed: %s", string(body))
	}

	var result struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", time.Time{}, err
	}

	return result.Token, result.ExpiresAt, nil
}

func (c *Client) Logout() error {
	resp, err := c.doRequest("POST", "/api/auth/logout", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status: %d", resp.StatusCode)
	}

	return auth.DeleteClientToken()
}

func (c *Client) ListPasswords() ([]*models.Password, error) {
	resp, err := c.doRequest("GET", "/api/passwords", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var passwords []*models.Password
	if err := json.NewDecoder(resp.Body).Decode(&passwords); err != nil {
		return nil, err
	}

	return passwords, nil
}

func (c *Client) GetPassword(id int64) (*models.Password, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/passwords/%d", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var password models.Password
	if err := json.NewDecoder(resp.Body).Decode(&password); err != nil {
		return nil, err
	}

	return &password, nil
}

func (c *Client) AddPassword(p *models.Password) error {
	resp, err := c.doRequest("POST", "/api/passwords", p)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(p)
}

func (c *Client) UpdatePassword(p *models.Password) error {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/passwords/%d", p.ID), p)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DeletePassword(id int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/passwords/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) SearchPasswords(query string) ([]*models.Password, error) {
	resp, err := c.doRequest("GET", "/api/passwords/search?q="+query, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var passwords []*models.Password
	if err := json.NewDecoder(resp.Body).Decode(&passwords); err != nil {
		return nil, err
	}

	return passwords, nil
}
