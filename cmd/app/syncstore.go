package main

import (
	"fmt"
	"sync"
	"time"
)

// SyncStore wraps LocalStore with synchronization capabilities
type SyncStore struct {
	local   *LocalStore
	client  *SyncClient
	config  SyncConfig
	mu      sync.RWMutex
	stopCh  chan struct{}
	running bool
}

// NewSyncStore creates a new sync store instance
func NewSyncStore(local *LocalStore, client *SyncClient, config SyncConfig) *SyncStore {
	return &SyncStore{
		local:   local,
		client:  client,
		config:  config,
		stopCh:  make(chan struct{}),
		running: false,
	}
}

// GetTodoLists retrieves all todo lists
func (s *SyncStore) GetTodoLists() ([]todoList, error) {
	return s.local.GetTodoLists()
}

// CreateTodoList creates a new todo list
func (s *SyncStore) CreateTodoList(name string) (int, error) {
	id, err := s.local.CreateTodoList(name)
	if err != nil {
		return 0, err
	}

	s.local.LogChange("list", id, "create")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return id, nil
}

// UpdateTodoListName updates a todo list name
func (s *SyncStore) UpdateTodoListName(id int, name string) error {
	err := s.local.UpdateTodoListName(id, name)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// DeleteTodoList deletes a todo list
func (s *SyncStore) DeleteTodoList(id int) error {
	err := s.local.DeleteTodoList(id)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "delete")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// ArchiveTodoList archives a todo list
func (s *SyncStore) ArchiveTodoList(id int) error {
	err := s.local.ArchiveTodoList(id)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// UnarchiveTodoList unarchives a todo list
func (s *SyncStore) UnarchiveTodoList(id int) error {
	err := s.local.UnarchiveTodoList(id)
	if err != nil {
		return err
	}

	s.local.LogChange("list", id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// GetItems retrieves all items
func (s *SyncStore) GetItems() ([]todoItem, error) {
	return s.local.GetItems()
}

// GetItemByID retrieves an item by ID
func (s *SyncStore) GetItemByID(id int) (todoItem, error) {
	return s.local.GetItemByID(id)
}

// GetItemByClientID retrieves an item by client ID
func (s *SyncStore) GetItemByClientID(clientID string) (todoItem, error) {
	return s.local.GetItemByClientID(clientID)
}

// SaveItem saves a new item
func (s *SyncStore) SaveItem(item todoItem) error {
	if item.clientID == "" {
		item.clientID = generateClientID()
	}

	err := s.local.SaveItem(item)
	if err != nil {
		return err
	}

	// Get the item back to get the ID
	savedItem, err := s.local.GetItemByClientID(item.clientID)
	if err == nil {
		s.local.LogChange("task", savedItem.id, "create")
	}

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// UpdateItem updates an existing item
func (s *SyncStore) UpdateItem(item todoItem) error {
	err := s.local.UpdateItem(item)
	if err != nil {
		return err
	}

	s.local.LogChange("task", item.id, "update")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// DeleteItem deletes an item
func (s *SyncStore) DeleteItem(id int) error {
	err := s.local.DeleteItem(id)
	if err != nil {
		return err
	}

	s.local.LogChange("task", id, "delete")

	// Trigger sync if enabled
	if s.config.AutoSyncOnChange && s.client.IsOnline() {
		go s.FullSync()
	}

	return nil
}

// GetLastSyncTime retrieves the last sync timestamp
func (s *SyncStore) GetLastSyncTime() (int64, error) {
	return s.local.GetLastSyncTime()
}

// SetLastSyncTime updates the last sync timestamp
func (s *SyncStore) SetLastSyncTime(timestamp int64) error {
	return s.local.SetLastSyncTime(timestamp)
}

// GetPendingChanges retrieves pending changes
func (s *SyncStore) GetPendingChanges() ([]Change, error) {
	return s.local.GetPendingChanges()
}

// MarkChangeSynced marks a change as synced
func (s *SyncStore) MarkChangeSynced(changeID int) error {
	return s.local.MarkChangeSynced(changeID)
}

// LogChange logs a change
func (s *SyncStore) LogChange(entityType string, entityID int, changeType string) error {
	return s.local.LogChange(entityType, entityID, changeType)
}

// FullSync performs a complete sync (pull then push)
func (s *SyncStore) FullSync() error {
	lastSync, _ := s.local.GetLastSyncTime()

	// Pull first to get latest state from server
	if err := s.PullChanges(lastSync); err != nil {
		return fmt.Errorf("pull changes failed: %w", err)
	}

	// Push local changes
	if err := s.PushChanges(); err != nil {
		return fmt.Errorf("push changes failed: %w", err)
	}

	// Update last sync time
	s.local.SetLastSyncTime(time.Now().Unix())

	return nil
}

// PullChanges pulls changes from server and applies them locally
func (s *SyncStore) PullChanges(since int64) error {
	resp, err := s.client.PullChanges(since)
	if err != nil {
		return err
	}

	if resp == nil {
		return nil
	}

	// Apply task changes
	for _, serverTask := range resp.Tasks {
		// Try to find existing local task by client ID
		localTask, err := s.local.GetItemByClientID(serverTask.ClientID)
		if err != nil {
			// Doesn't exist locally - create it
			newItem := todoItem{
				clientID:      serverTask.ClientID,
				serverID:      0,
				done:          serverTask.Done,
				todo:          serverTask.Todo,
				priority:      serverTask.Priority,
				dateCompleted: serverTask.DateCompleted,
				dateAdded:     serverTask.DateAdded,
				dueDate:       serverTask.DueDate,
				deleted:       serverTask.Deleted,
				deletedAt:     serverTask.DeletedAt,
				todoListID:    serverTask.TodoListID,
				version:       serverTask.Version,
			}
			s.local.SaveItem(newItem)
			continue
		}

		// Task exists - resolve conflict using "last write wins"
		if serverTask.UpdatedAt > localTask.dateAdded {
			// Server is newer - update local copy
			localTask.done = serverTask.Done
			localTask.todo = serverTask.Todo
			localTask.priority = serverTask.Priority
			localTask.dateCompleted = serverTask.DateCompleted
			localTask.dueDate = serverTask.DueDate
			localTask.deleted = serverTask.Deleted
			localTask.deletedAt = serverTask.DeletedAt
			localTask.todoListID = serverTask.TodoListID
			localTask.version = serverTask.Version
			s.local.UpdateItem(localTask)
		}
		// Otherwise local is newer, leave it as is
	}

	// Apply list changes
	for _, serverList := range resp.Lists {
		// For now, we'll skip list syncing as it requires more complex logic
		// with archived lists
		_ = serverList
	}

	return nil
}

// PushChanges pushes pending local changes to the server
func (s *SyncStore) PushChanges() error {
	items, err := s.local.GetItems()
	if err != nil {
		return err
	}

	lists, err := s.local.GetTodoLists()
	if err != nil {
		return err
	}

	if err := s.client.PushChanges(items, lists); err != nil {
		return err
	}

	// Mark all changes as synced
	changes, err := s.local.GetPendingChanges()
	if err != nil {
		return err
	}

	for _, change := range changes {
		s.local.MarkChangeSynced(change.id)
	}

	return nil
}

// StartBackgroundSync starts the background sync goroutine
func (s *SyncStore) StartBackgroundSync() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Duration(s.config.SyncIntervalSeconds) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if s.client.IsOnline() {
					s.FullSync()
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

// StopBackgroundSync stops the background sync goroutine
func (s *SyncStore) StopBackgroundSync() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		close(s.stopCh)
		s.running = false
	}
}
