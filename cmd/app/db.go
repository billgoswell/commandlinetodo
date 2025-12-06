package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"time"
)

type todoitem struct {
	id            int
	done          bool
	todo          string
	priority      int
	datecompleted int64
	dateadded     int64
	duedate       int64
	deleted       bool
	deletedat     int64
	todolistID    int
}

var db *sql.DB

func initDB() (*sql.DB, error) {
	var err error
	db, err = sql.Open("sqlite", "./todo.db")
	if err != nil {
		fmt.Println("Failed to open DB:", err)
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
	// Create todolists table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS todolists (
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

	// Create tasks table with todolist_id foreign key
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo TEXT NOT NULL,
		priority INTEGER DEFAULT 4,
		done BOOLEAN DEFAULT 0,
		dateadded INTEGER,
		datecompleted INTEGER DEFAULT 0,
		duedate INTEGER DEFAULT 0,
		deleted BOOLEAN DEFAULT 0,
		deletedat INTEGER DEFAULT 0,
		todolist_id INTEGER DEFAULT 1
	)`)
	return err
}

func getItemsFromDB() ([]todoitem, error) {
	rows, err := db.Query("SELECT id, todo, priority, done, dateadded, datecompleted, duedate, deleted, deletedat, todolist_id FROM tasks WHERE deleted = 0 ORDER BY id")
	if err != nil {
		fmt.Println("Failed to query items:", err)
		return []todoitem{}, err
	}
	defer rows.Close()

	items := []todoitem{}
	for rows.Next() {
		var item todoitem
		if err := rows.Scan(&item.id, &item.todo, &item.priority, &item.done, &item.dateadded, &item.datecompleted, &item.duedate, &item.deleted, &item.deletedat, &item.todolistID); err != nil {
			fmt.Println("Failed to scan item:", err)
			return []todoitem{}, err
		}
		items = append(items, item)
	}

	return items, nil
}

func saveItemToDB(item todoitem) error {
	_, err := db.Exec(
		"INSERT INTO tasks (todo, priority, done, dateadded, duedate, deleted, todolist_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		item.todo, item.priority, item.done, time.Now().Unix(), item.duedate, 0, item.todolistID,
	)
	if err != nil {
		fmt.Println("Failed to insert item:", err)
		return err
	}
	return nil
}

func updateItemInDB(item todoitem) error {
	_, err := db.Exec(
		"UPDATE tasks SET todo = ?, done = ?, priority = ?, datecompleted = ?, duedate = ? WHERE id = ?",
		item.todo, item.done, item.priority, item.datecompleted, item.duedate, item.id,
	)
	if err != nil {
		fmt.Println("Failed to update item:", err)
		return err
	}
	return nil
}

func markItemAsDeleted(id int) error {
	_, err := db.Exec(
		"UPDATE tasks SET deleted = 1, deletedat = ? WHERE id = ?",
		time.Now().Unix(), id,
	)
	if err != nil {
		fmt.Println("Failed to delete item:", err)
		return err
	}
	return nil
}

// Todolist management functions

func getTodoLists() ([]todolist, error) {
	rows, err := db.Query("SELECT id, name, display_order, archived, created_at, updated_at FROM todolists WHERE archived = 0 ORDER BY display_order")
	if err != nil {
		fmt.Println("Failed to query todolists:", err)
		return []todolist{}, err
	}
	defer rows.Close()

	lists := []todolist{}
	for rows.Next() {
		var list todolist
		if err := rows.Scan(&list.id, &list.name, &list.displayOrder, &list.archived, &list.createdAt, &list.updatedAt); err != nil {
			fmt.Println("Failed to scan todolist:", err)
			return []todolist{}, err
		}
		lists = append(lists, list)
	}

	return lists, nil
}

func createTodoList(name string) (int, error) {
	now := time.Now().Unix()
	result, err := db.Exec(
		"INSERT INTO todolists (name, display_order, archived, created_at, updated_at) VALUES (?, (SELECT COUNT(*) FROM todolists), 0, ?, ?)",
		name, now, now,
	)
	if err != nil {
		fmt.Println("Failed to create todolist:", err)
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
		"UPDATE todolists SET name = ?, updated_at = ? WHERE id = ?",
		name, now, id,
	)
	if err != nil {
		fmt.Println("Failed to update todolist name:", err)
		return err
	}
	return nil
}

func deleteTodoList(id int) error {
	// Mark list as archived instead of deleting
	return archiveTodoList(id)
}

func archiveTodoList(id int) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"UPDATE todolists SET archived = 1, updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		fmt.Println("Failed to archive todolist:", err)
		return err
	}
	return nil
}

func unarchiveTodoList(id int) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		"UPDATE todolists SET archived = 0, updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		fmt.Println("Failed to unarchive todolist:", err)
		return err
	}
	return nil
}

func fixExistingTaskListIDs() error {
	// Set all tasks with todolist_id = 0 or NULL to 1 (General list)
	_, err := db.Exec("UPDATE tasks SET todolist_id = 1 WHERE todolist_id IS NULL OR todolist_id = 0")
	if err != nil {
		fmt.Println("Warning: failed to fix task list IDs:", err)
		return err
	}
	return nil
}
