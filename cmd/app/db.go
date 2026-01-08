package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"time"
	"github.com/google/uuid"
)

type todoItem struct {
	id            int
	clientID      string // UUID for offline-first sync
	serverID      int    // Server's ID (0 if not synced yet)
	done          bool
	todo          string
	priority      int
	dateCompleted int64
	dateAdded     int64
	dueDate       int64
	deleted       bool
	deletedAt     int64
	todoListID    int
	version       int // For conflict detection
}

var db *sql.DB

func logError(operation string, err error) {
	fmt.Printf("Failed to %s: %v\n", operation, err)
}

func now() int64 {
	return time.Now().Unix()
}

func executeStmt(operation string, query string, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	if err != nil {
		logError(operation, err)
		return err
	}
	return nil
}

func executeStmtWithID(operation string, query string, args ...interface{}) (int, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		logError(operation, err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logError(operation, err)
		return 0, err
	}

	return int(id), nil
}

func initDB(dbPath string) (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		logError("open database at "+dbPath, err)
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if err := executeStmt("enable foreign keys", "PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	if err := createTableIfNotExists(); err != nil {
		logError("create table", err)
		return nil, err
	}

	if err := createSyncTables(); err != nil {
		logError("create sync tables", err)
		return nil, err
	}

	if err := migrateSyncColumns(); err != nil {
		logError("migrate sync columns", err)
		return nil, err
	}

	if err := fixExistingTaskListIDs(); err != nil {
		fmt.Println("Warning: failed to fix task list IDs:", err)
	}

	return db, nil
}

func createTableIfNotExists() error {
	if err := executeStmt("create todoLists table", `CREATE TABLE IF NOT EXISTS todoLists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		display_order INTEGER DEFAULT 0,
		archived BOOLEAN DEFAULT 0,
		created_at INTEGER,
		updated_at INTEGER
	)`); err != nil {
		return err
	}

	return executeStmt("create tasks table", `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo TEXT NOT NULL,
		priority INTEGER DEFAULT 4,
		done BOOLEAN DEFAULT 0,
		dateAdded INTEGER,
		dateCompleted INTEGER DEFAULT 0,
		dueDate INTEGER DEFAULT 0,
		deleted BOOLEAN DEFAULT 0,
		deletedAt INTEGER DEFAULT 0,
		todoList_id INTEGER DEFAULT 1,
		FOREIGN KEY (todoList_id) REFERENCES todoLists(id)
	)`)
}

func getItemsFromDB() ([]todoItem, error) {
	rows, err := db.Query("SELECT id, todo, priority, done, dateAdded, dateCompleted, dueDate, deleted, deletedAt, todoList_id, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1) FROM tasks WHERE deleted = 0 ORDER BY id")
	if err != nil {
		fmt.Println("Failed to query items:", err)
		return []todoItem{}, err
	}
	defer rows.Close()

	items := []todoItem{}
	for rows.Next() {
		var item todoItem
		if err := rows.Scan(&item.id, &item.todo, &item.priority, &item.done, &item.dateAdded, &item.dateCompleted, &item.dueDate, &item.deleted, &item.deletedAt, &item.todoListID, &item.clientID, &item.serverID, &item.version); err != nil {
			fmt.Println("Failed to scan item:", err)
			return []todoItem{}, err
		}
		// Validate priority to ensure it's a valid value (1-4)
		item.priority = validatePriority(item.priority)
		// Generate client ID if missing (for backward compatibility)
		if item.clientID == "" {
			item.clientID = generateClientID()
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating items:", err)
		return []todoItem{}, err
	}

	return items, nil
}

func saveItemToDB(item todoItem) error {
	return executeStmt("insert item",
		"INSERT INTO tasks (todo, priority, done, dateAdded, dueDate, deleted, todoList_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		item.todo, item.priority, item.done, now(), item.dueDate, 0, item.todoListID,
	)
}

func updateItemInDB(item todoItem) error {
	return executeStmt("update item",
		"UPDATE tasks SET todo = ?, done = ?, priority = ?, dateCompleted = ?, dueDate = ?, todoList_id = ? WHERE id = ?",
		item.todo, item.done, item.priority, item.dateCompleted, item.dueDate, item.todoListID, item.id,
	)
}

func markItemAsDeleted(id int) error {
	return executeStmt("delete item",
		"UPDATE tasks SET deleted = 1, deletedAt = ? WHERE id = ?",
		now(), id,
	)
}

func getTodoLists() ([]todoList, error) {
	rows, err := db.Query("SELECT id, name, display_order, archived, created_at, updated_at, COALESCE(client_id, ''), COALESCE(server_id, 0), COALESCE(version, 1) FROM todoLists WHERE archived = 0 ORDER BY display_order")
	if err != nil {
		fmt.Println("Failed to query todoLists:", err)
		return []todoList{}, err
	}
	defer rows.Close()

	lists := []todoList{}
	for rows.Next() {
		var list todoList
		if err := rows.Scan(&list.id, &list.name, &list.displayOrder, &list.archived, &list.createdAt, &list.updatedAt, &list.clientID, &list.serverID, &list.version); err != nil {
			fmt.Println("Failed to scan todoList:", err)
			return []todoList{}, err
		}
		// Generate client ID if missing (for backward compatibility)
		if list.clientID == "" {
			list.clientID = generateClientID()
		}
		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating todoLists:", err)
		return []todoList{}, err
	}

	return lists, nil
}

func createTodoList(name string) (int, error) {
	return executeStmtWithID("create todo list",
		"INSERT INTO todoLists (name, display_order, archived, created_at, updated_at) VALUES (?, (SELECT COUNT(*) FROM todoLists), 0, ?, ?)",
		name, now(), now(),
	)
}

func updateTodoListName(id int, name string) error {
	return executeStmt("update todo list name",
		"UPDATE todoLists SET name = ?, updated_at = ? WHERE id = ?",
		name, now(), id,
	)
}

func deleteTodoList(id int) error {
	tx, err := db.Begin()
	if err != nil {
		logError("begin transaction", err)
		return err
	}
	defer tx.Rollback()

	timestamp := now()

	_, err = tx.Exec(
		"UPDATE todoLists SET archived = 1, updated_at = ? WHERE id = ?",
		timestamp, id,
	)
	if err != nil {
		logError("archive todo list", err)
		return err
	}

	_, err = tx.Exec(
		"UPDATE tasks SET deleted = 1, deletedAt = ? WHERE todoList_id = ? AND deleted = 0",
		timestamp, id,
	)
	if err != nil {
		logError("delete tasks in list", err)
		return err
	}

	return tx.Commit()
}

func setTodoListArchived(id int, archived bool) error {
	operation := "archive todo list"
	if !archived {
		operation = "unarchive todo list"
	}

	return executeStmt(operation,
		"UPDATE todoLists SET archived = ?, updated_at = ? WHERE id = ?",
		archived, now(), id,
	)
}

func archiveTodoList(id int) error {
	return setTodoListArchived(id, true)
}

func unarchiveTodoList(id int) error {
	return setTodoListArchived(id, false)
}

func fixExistingTaskListIDs() error {
	if err := executeStmt("fix task list IDs",
		"UPDATE tasks SET todoList_id = 1 WHERE todoList_id IS NULL OR todoList_id = 0",
	); err != nil {
		return err
	}

	return verifyTaskListIDs()
}

func verifyTaskListIDs() error {
	var orphanedCount int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE deleted = 0 AND (todoList_id IS NULL OR todoList_id = 0)").Scan(&orphanedCount)
	if err != nil {
		logError("verify task list IDs", err)
		return err
	}

	if orphanedCount > 0 {
		return fmt.Errorf("data integrity error: found %d orphaned tasks with invalid list IDs", orphanedCount)
	}

	return nil
}

// Sync infrastructure functions

func generateClientID() string {
	return uuid.New().String()
}

func createSyncTables() error {
	// Create sync metadata table
	if err := executeStmt("create sync_metadata table", `CREATE TABLE IF NOT EXISTS sync_metadata (
		key TEXT PRIMARY KEY,
		value TEXT
	)`); err != nil {
		return err
	}

	// Create change log table
	return executeStmt("create change_log table", `CREATE TABLE IF NOT EXISTS change_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		change_type TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		synced BOOLEAN DEFAULT 0
	)`)
}

func columnExists(tableName, columnName string) (bool, error) {
	var name string
	err := db.QueryRow(
		"SELECT name FROM pragma_table_info(?) WHERE name = ?",
		tableName, columnName,
	).Scan(&name)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func migrateSyncColumns() error {
	// Check and add sync columns to todoLists
	if exists, err := columnExists("todoLists", "client_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add client_id to todoLists",
			"ALTER TABLE todoLists ADD COLUMN client_id TEXT"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("todoLists", "server_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add server_id to todoLists",
			"ALTER TABLE todoLists ADD COLUMN server_id INTEGER DEFAULT 0"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("todoLists", "version"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add version to todoLists",
			"ALTER TABLE todoLists ADD COLUMN version INTEGER DEFAULT 1"); err != nil {
			return err
		}
	}

	// Check and add sync columns to tasks
	if exists, err := columnExists("tasks", "client_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add client_id to tasks",
			"ALTER TABLE tasks ADD COLUMN client_id TEXT"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("tasks", "server_id"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add server_id to tasks",
			"ALTER TABLE tasks ADD COLUMN server_id INTEGER DEFAULT 0"); err != nil {
			return err
		}
	}

	if exists, err := columnExists("tasks", "version"); err != nil {
		return err
	} else if !exists {
		if err := executeStmt("add version to tasks",
			"ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1"); err != nil {
			return err
		}
	}

	return nil
}

func logChange(entityType string, entityID int, changeType string) error {
	return executeStmt("log change",
		"INSERT INTO change_log (entity_type, entity_id, change_type, timestamp, synced) VALUES (?, ?, ?, ?, 0)",
		entityType, entityID, changeType, now(),
	)
}

func getPendingChanges() ([]Change, error) {
	rows, err := db.Query("SELECT id, entity_type, entity_id, change_type, timestamp, synced FROM change_log WHERE synced = 0 ORDER BY timestamp")
	if err != nil {
		logError("query pending changes", err)
		return []Change{}, err
	}
	defer rows.Close()

	changes := []Change{}
	for rows.Next() {
		var change Change
		if err := rows.Scan(&change.id, &change.entityType, &change.entityID, &change.changeType, &change.timestamp, &change.synced); err != nil {
			logError("scan change", err)
			return []Change{}, err
		}
		changes = append(changes, change)
	}

	if err := rows.Err(); err != nil {
		logError("iterate changes", err)
		return []Change{}, err
	}

	return changes, nil
}

func markChangeSynced(changeID int) error {
	return executeStmt("mark change synced",
		"UPDATE change_log SET synced = 1 WHERE id = ?",
		changeID,
	)
}

func getLastSyncTime() (int64, error) {
	var timestamp int64
	err := db.QueryRow("SELECT value FROM sync_metadata WHERE key = 'last_sync_time'").Scan(&timestamp)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		logError("get last sync time", err)
		return 0, err
	}

	return timestamp, nil
}

func setLastSyncTime(timestamp int64) error {
	// Use INSERT OR REPLACE to handle both insert and update
	return executeStmt("set last sync time",
		"INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_sync_time', ?)",
		fmt.Sprintf("%d", timestamp),
	)
}
