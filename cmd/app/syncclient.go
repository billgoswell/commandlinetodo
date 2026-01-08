package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ConnectivityStatus represents the current connection status
type ConnectivityStatus int

const (
	StatusOffline ConnectivityStatus = iota
	StatusOnline
	StatusSyncing
	StatusError
)

// SyncClient handles HTTP communication with the sync server
type SyncClient struct {
	baseURL      string
	httpClient   *http.Client
	apiKey       string
	deviceID     string
	status       ConnectivityStatus
	lastCheck    time.Time
	lastError    error
	mu           sync.RWMutex
	checkTimeout time.Duration
}

// TaskPayload represents a task for sync
type TaskPayload struct {
	ClientID      string `json:"client_id"`
	Todo          string `json:"todo"`
	Priority      int    `json:"priority"`
	Done          bool   `json:"done"`
	DateAdded     int64  `json:"date_added"`
	DateCompleted int64  `json:"date_completed"`
	DueDate       int64  `json:"due_date"`
	Deleted       bool   `json:"deleted"`
	DeletedAt     int64  `json:"deleted_at"`
	TodoListID    int    `json:"todo_list_id"`
	UpdatedAt     int64  `json:"updated_at"`
	Version       int    `json:"version"`
}

// ListPayload represents a todo list for sync
type ListPayload struct {
	ClientID    string `json:"client_id"`
	Name        string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	Archived    bool   `json:"archived"`
	UpdatedAt   int64  `json:"updated_at"`
	Version     int    `json:"version"`
}

// PullRequest is the request for pulling changes
type PullRequest struct {
	Since int64 `json:"since"`
}

// PullResponse contains the changes from the server
type PullResponse struct {
	Tasks []TaskPayload `json:"tasks"`
	Lists []ListPayload `json:"lists"`
}

// PushRequest contains changes to push to server
type PushRequest struct {
	Tasks []TaskPayload `json:"tasks"`
	Lists []ListPayload `json:"lists"`
}

// HealthResponse is the response from the health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// NewSyncClient creates a new sync client
func NewSyncClient(config SyncConfig) *SyncClient {
	return &SyncClient{
		baseURL:      config.ServerURL,
		apiKey:       config.APIKey,
		deviceID:     config.DeviceID,
		status:       StatusOffline,
		lastCheck:    time.Time{},
		checkTimeout: 5 * time.Second,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
	}
}

// CheckConnectivity checks if the server is reachable
func (c *SyncClient) CheckConnectivity() ConnectivityStatus {
	c.mu.RLock()
	// Return cached status if checked recently (within 5 seconds)
	if time.Since(c.lastCheck) < c.checkTimeout {
		defer c.mu.RUnlock()
		return c.status
	}
	c.mu.RUnlock()

	// Try to reach the health endpoint
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		c.updateStatus(StatusError, err)
		return StatusError
	}

	c.addAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.updateStatus(StatusOffline, err)
		return StatusOffline
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.updateStatus(StatusOnline, nil)
		return StatusOnline
	}

	c.updateStatus(StatusError, fmt.Errorf("server returned status %d", resp.StatusCode))
	return StatusError
}

// IsOnline returns true if the sync client is online
func (c *SyncClient) IsOnline() bool {
	return c.CheckConnectivity() == StatusOnline
}

// PullChanges retrieves changes from the server since the given timestamp
func (c *SyncClient) PullChanges(since int64) (*PullResponse, error) {
	if !c.IsOnline() {
		return nil, fmt.Errorf("not connected to sync server")
	}

	pullReq := PullRequest{Since: since}
	body, err := json.Marshal(pullReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/sync/pull", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	c.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pull changes failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var pullResp PullResponse
	if err := json.NewDecoder(resp.Body).Decode(&pullResp); err != nil {
		return nil, err
	}

	return &pullResp, nil
}

// PushChanges sends local changes to the server
func (c *SyncClient) PushChanges(items []todoItem, lists []todoList) error {
	if !c.IsOnline() {
		return fmt.Errorf("not connected to sync server")
	}

	// Convert items to payloads
	taskPayloads := make([]TaskPayload, len(items))
	for i, item := range items {
		taskPayloads[i] = TaskPayload{
			ClientID:      item.clientID,
			Todo:          item.todo,
			Priority:      item.priority,
			Done:          item.done,
			DateAdded:     item.dateAdded,
			DateCompleted: item.dateCompleted,
			DueDate:       item.dueDate,
			Deleted:       item.deleted,
			DeletedAt:     item.deletedAt,
			TodoListID:    item.todoListID,
			UpdatedAt:     item.dateAdded, // TODO: Use actual updated timestamp
			Version:       item.version,
		}
	}

	// Convert lists to payloads
	listPayloads := make([]ListPayload, len(lists))
	for i, list := range lists {
		listPayloads[i] = ListPayload{
			ClientID:    list.clientID,
			Name:        list.name,
			DisplayOrder: list.displayOrder,
			Archived:    list.archived,
			UpdatedAt:   list.updatedAt,
			Version:     list.version,
		}
	}

	pushReq := PushRequest{
		Tasks: taskPayloads,
		Lists: listPayloads,
	}

	body, err := json.Marshal(pushReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/sync/push", bytes.NewReader(body))
	if err != nil {
		return err
	}

	c.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push changes failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// addAuthHeaders adds authentication headers to requests
func (c *SyncClient) addAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Device-ID", c.deviceID)
}

// updateStatus updates the connection status
func (c *SyncClient) updateStatus(status ConnectivityStatus, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = status
	c.lastError = err
	c.lastCheck = time.Now()
}

// GetLastError returns the last error that occurred
func (c *SyncClient) GetLastError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastError
}
