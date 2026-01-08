package main

import (
	"database/sql"
	"fmt"
)

// DataStore interface abstracts all data access operations
type DataStore interface {
	// Lists
	GetTodoLists() ([]todoList, error)
	CreateTodoList(name string) (int, error)
	UpdateTodoListName(id int, name string) error
	DeleteTodoList(id int) error
	ArchiveTodoList(id int) error
	UnarchiveTodoList(id int) error

	// Tasks
	GetItems() ([]todoItem, error)
	GetItemByID(id int) (todoItem, error)
	GetItemByClientID(clientID string) (todoItem, error)
	SaveItem(item todoItem) error
	UpdateItem(item todoItem) error
	DeleteItem(id int) error

	// Sync metadata
	GetLastSyncTime() (int64, error)
	SetLastSyncTime(timestamp int64) error
	GetPendingChanges() ([]Change, error)
	MarkChangeSynced(changeID int) error
	LogChange(entityType string, entityID int, changeType string) error
}

// LocalStore implements DataStore for local SQLite database
type LocalStore struct {
	db *sql.DB
}

// NewLocalStore creates a new local store instance
func NewLocalStore(database *sql.DB) *LocalStore {
	return &LocalStore{
		db: database,
	}
}

// GetTodoLists retrieves all non-archived todo lists
func (s *LocalStore) GetTodoLists() ([]todoList, error) {
	return getTodoLists()
}

// CreateTodoList creates a new todo list
func (s *LocalStore) CreateTodoList(name string) (int, error) {
	return createTodoList(name)
}

// UpdateTodoListName updates the name of a todo list
func (s *LocalStore) UpdateTodoListName(id int, name string) error {
	return updateTodoListName(id, name)
}

// DeleteTodoList deletes a todo list and all its tasks
func (s *LocalStore) DeleteTodoList(id int) error {
	return deleteTodoList(id)
}

// ArchiveTodoList archives a todo list
func (s *LocalStore) ArchiveTodoList(id int) error {
	return archiveTodoList(id)
}

// UnarchiveTodoList unarchives a todo list
func (s *LocalStore) UnarchiveTodoList(id int) error {
	return unarchiveTodoList(id)
}

// GetItems retrieves all non-deleted items
func (s *LocalStore) GetItems() ([]todoItem, error) {
	return getItemsFromDB()
}

// GetItemByID retrieves a single item by ID
func (s *LocalStore) GetItemByID(id int) (todoItem, error) {
	items, err := getItemsFromDB()
	if err != nil {
		return todoItem{}, err
	}

	for _, item := range items {
		if item.id == id {
			return item, nil
		}
	}

	return todoItem{}, fmt.Errorf("item not found: %d", id)
}

// GetItemByClientID retrieves an item by its client ID
func (s *LocalStore) GetItemByClientID(clientID string) (todoItem, error) {
	rows, err := s.db.Query(
		"SELECT id, todo, priority, done, dateAdded, dateCompleted, dueDate, deleted, deletedAt, todoList_id, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1) FROM tasks WHERE client_id = ? LIMIT 1",
		clientID,
	)
	if err != nil {
		logError("query item by client_id", err)
		return todoItem{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var item todoItem
		if err := rows.Scan(&item.id, &item.todo, &item.priority, &item.done, &item.dateAdded, &item.dateCompleted, &item.dueDate, &item.deleted, &item.deletedAt, &item.todoListID, &item.clientID, &item.serverID, &item.version); err != nil {
			logError("scan item by client_id", err)
			return todoItem{}, err
		}
		return item, nil
	}

	return todoItem{}, fmt.Errorf("item not found with client_id: %s", clientID)
}

// SaveItem saves a new item to the database
func (s *LocalStore) SaveItem(item todoItem) error {
	// Generate client ID if not present
	if item.clientID == "" {
		item.clientID = generateClientID()
	}
	return saveItemToDB(item)
}

// UpdateItem updates an existing item
func (s *LocalStore) UpdateItem(item todoItem) error {
	return updateItemInDB(item)
}

// DeleteItem marks an item as deleted
func (s *LocalStore) DeleteItem(id int) error {
	return markItemAsDeleted(id)
}

// GetLastSyncTime retrieves the timestamp of the last successful sync
func (s *LocalStore) GetLastSyncTime() (int64, error) {
	return getLastSyncTime()
}

// SetLastSyncTime updates the last sync timestamp
func (s *LocalStore) SetLastSyncTime(timestamp int64) error {
	return setLastSyncTime(timestamp)
}

// GetPendingChanges retrieves all unsynced changes
func (s *LocalStore) GetPendingChanges() ([]Change, error) {
	return getPendingChanges()
}

// MarkChangeSynced marks a change as successfully synced
func (s *LocalStore) MarkChangeSynced(changeID int) error {
	return markChangeSynced(changeID)
}

// LogChange records a local change for later sync
func (s *LocalStore) LogChange(entityType string, entityID int, changeType string) error {
	return logChange(entityType, entityID, changeType)
}
