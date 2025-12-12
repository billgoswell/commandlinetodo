package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"time"
)

type todoItem struct {
	id            int
	done          bool
	todo          string
	priority      int
	dateCompleted int64
	dateAdded     int64
	dueDate       int64
	deleted       bool
	deletedAt     int64
	todoListID    int
}

var db *sql.DB

func initDB() (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite", "./todo.db")
	if err != nil {
		fmt.Println("Failed to open DB:", err)
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		fmt.Println("Failed to enable foreign keys:", err)
		return nil, err
	}

	if err := createTableIfNotExists(); err != nil {
		fmt.Println("Failed to create table:", err)
		return nil, err
	}

	if err := fixExistingTaskListIDs(); err != nil {
		fmt.Println("Warning: failed to fix task list IDs:", err)
	}

	return db, nil
}

func createTableIfNotExists() error {
	// Create todoLists table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS todoLists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		display_order INTEGER DEFAULT 0,
		archived BOOLEAN DEFAULT 0,
		created_at INTEGER,
		updated_at INTEGER
	)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
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
	return err
}

func getItemsFromDB() ([]todoItem, error) {
	rows, err := db.Query("SELECT id, todo, priority, done, dateAdded, dateCompleted, dueDate, deleted, deletedAt, todoList_id FROM tasks WHERE deleted = 0 ORDER BY id")
	if err != nil {
		fmt.Println("Failed to query items:", err)
		return []todoItem{}, err
	}
	defer rows.Close()

	items := []todoItem{}
	for rows.Next() {
		var item todoItem
		if err := rows.Scan(&item.id, &item.todo, &item.priority, &item.done, &item.dateAdded, &item.dateCompleted, &item.dueDate, &item.deleted, &item.deletedAt, &item.todoListID); err != nil {
			fmt.Println("Failed to scan item:", err)
			return []todoItem{}, err
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
	_, err := db.Exec(
		"INSERT INTO tasks (todo, priority, done, dateAdded, dueDate, deleted, todoList_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		item.todo, item.priority, item.done, time.Now().Unix(), item.dueDate, 0, item.todoListID,
	)
	if err != nil {
		fmt.Println("Failed to insert item:", err)
		return err
	}
	return nil
}

func updateItemInDB(item todoItem) error {
	_, err := db.Exec(
		"UPDATE tasks SET todo = ?, done = ?, priority = ?, dateCompleted = ?, dueDate = ?, todoList_id = ? WHERE id = ?",
		item.todo, item.done, item.priority, item.dateCompleted, item.dueDate, item.todoListID, item.id,
	)
	if err != nil {
		fmt.Println("Failed to update item:", err)
		return err
	}
	return nil
}

func markItemAsDeleted(id int) error {
	_, err := db.Exec(
		"UPDATE tasks SET deleted = 1, deletedAt = ? WHERE id = ?",
		time.Now().Unix(), id,
	)
	if err != nil {
		fmt.Println("Failed to delete item:", err)
		return err
	}
	return nil
}

// Todolist management functions

func getTodoLists() ([]todoList, error) {
	rows, err := db.Query("SELECT id, name, display_order, archived, created_at, updated_at FROM todoLists WHERE archived = 0 ORDER BY display_order")
	if err != nil {
		fmt.Println("Failed to query todoLists:", err)
		return []todoList{}, err
	}
	defer rows.Close()

	lists := []todoList{}
	for rows.Next() {
		var list todoList
		if err := rows.Scan(&list.id, &list.name, &list.displayOrder, &list.archived, &list.createdAt, &list.updatedAt); err != nil {
			fmt.Println("Failed to scan todoList:", err)
			return []todoList{}, err
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
	now := time.Now().Unix()
	result, err := db.Exec(
		"INSERT INTO todoLists (name, display_order, archived, created_at, updated_at) VALUES (?, (SELECT COUNT(*) FROM todoLists), 0, ?, ?)",
		name, now, now,
	)
	if err != nil {
		fmt.Println("Failed to create todoList:", err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Failed to get last insert id:", err)
		return 0, err
	}

	return int(id), nil
}

func updateTodoListName(id int, name string) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"UPDATE todoLists SET name = ?, updated_at = ? WHERE id = ?",
		name, now, id,
	)
	if err != nil {
		fmt.Println("Failed to update todoList name:", err)
		return err
	}
	return nil
}

func deleteTodoList(id int) error {
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("Failed to begin transaction:", err)
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Archive the list
	_, err = tx.Exec(
		"UPDATE todoLists SET archived = 1, updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		fmt.Println("Failed to archive todoList:", err)
		return err
	}

	// Mark all tasks in this list as deleted
	_, err = tx.Exec(
		"UPDATE tasks SET deleted = 1, deletedAt = ? WHERE todoList_id = ? AND deleted = 0",
		now, id,
	)
	if err != nil {
		fmt.Println("Failed to delete tasks in list:", err)
		return err
	}

	return tx.Commit()
}

func archiveTodoList(id int) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"UPDATE todoLists SET archived = 1, updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		fmt.Println("Failed to archive todoList:", err)
		return err
	}
	return nil
}

func unarchiveTodoList(id int) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"UPDATE todoLists SET archived = 0, updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		fmt.Println("Failed to unarchive todoList:", err)
		return err
	}
	return nil
}

func fixExistingTaskListIDs() error {
	// Set all tasks with todoList_id = 0 or NULL to 1 (General list)
	_, err := db.Exec("UPDATE tasks SET todoList_id = 1 WHERE todoList_id IS NULL OR todoList_id = 0")
	if err != nil {
		fmt.Println("Warning: failed to fix task list IDs:", err)
		return err
	}
	return nil
}
