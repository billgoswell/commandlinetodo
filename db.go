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

	return db, nil
}

func createTableIfNotExists() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo TEXT NOT NULL,
		priority INTEGER DEFAULT 4,
		done BOOLEAN DEFAULT 0,
		dateadded INTEGER,
		datecompleted INTEGER DEFAULT 0,
		duedate INTEGER DEFAULT 0,
		deleted BOOLEAN DEFAULT 0,
		deletedat INTEGER DEFAULT 0
	)`)
	return err
}

func getItemsFromDB() ([]todoitem, error) {
	rows, err := db.Query("SELECT id, todo, priority, done, dateadded, datecompleted, duedate, deleted, deletedat FROM tasks WHERE deleted = 0 ORDER BY id")
	if err != nil {
		fmt.Println("Failed to query items:", err)
		return []todoitem{}, err
	}
	defer rows.Close()

	items := []todoitem{}
	for rows.Next() {
		var item todoitem
		if err := rows.Scan(&item.id, &item.todo, &item.priority, &item.done, &item.dateadded, &item.datecompleted, &item.duedate, &item.deleted, &item.deletedat); err != nil {
			fmt.Println("Failed to scan item:", err)
			return []todoitem{}, err
		}
		items = append(items, item)
	}

	return items, nil
}

func saveItemToDB(item todoitem) error {
	_, err := db.Exec(
		"INSERT INTO tasks (todo, priority, done, dateadded, duedate, deleted) VALUES (?, ?, ?, ?, ?, ?)",
		item.todo, item.priority, item.done, time.Now().Unix(), item.duedate, 0,
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
