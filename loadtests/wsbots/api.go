package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *APIClient) Register(email, username, password string) error {
	status, _, err := c.post("/auth/register", map[string]string{
		"email":    email,
		"username": username,
		"password": password,
	}, "")
	if err != nil {
		return err
	}
	if status == http.StatusCreated {
		return nil
	}
	if status == http.StatusConflict {
		return nil
	}
	return fmt.Errorf("register %s: status %d", email, status)
}

func (c *APIClient) Login(email, password string) (string, error) {
	status, body, err := c.post("/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("login %s: status %d body=%s", email, status, string(body))
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("login %s: empty access_token", email)
	}
	return out.AccessToken, nil
}

func (c *APIClient) SearchUser(token, username string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/users/search?username="+url.QueryEscape(username), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search %s: status %d body=%s", username, res.StatusCode, string(body))
	}

	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *APIClient) CreateConversation(token, participantID string) (string, error) {
	status, body, err := c.post("/conversations", map[string]string{
		"participant_id": participantID,
	}, token)
	if err != nil {
		return "", err
	}
	if status != http.StatusCreated {
		return "", fmt.Errorf("create conversation: status %d body=%s", status, string(body))
	}

	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *APIClient) GetWSTicket(token string) (string, error) {
	status, body, err := c.post("/ws/ticket", map[string]string{}, token)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("ws ticket: status %d body=%s", status, string(body))
	}
	var out struct {
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.Ticket == "" {
		return "", fmt.Errorf("ws ticket: empty ticket")
	}
	return out.Ticket, nil
}

func (c *APIClient) post(path string, payload any, token string) (int, []byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, err
	}
	return res.StatusCode, body, nil
}
