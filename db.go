package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

type todoitem struct {
	id            int
	done          bool
	todo          string
	priority      int
	notes         string
	datecompleted int
	dateadded     int
}

func openDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./todo.db")
	if err != nil {
		fmt.Println("Failed to Open DB")
	}
	if checkTable(db) {
		createTable(db)
	}
	return db
}

func checkTable(db *sql.DB) bool {
	if _, err := db.Query("SELECT * FROM tasks"); err == nil {
		return true
	}
	return false
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE "tasks" ( "id" INTEGER, "todo" TEXT NOT NULL, priority INTEGER, "notes" TEXT, "done" BOOLEAN, "dateadded")`)
	return err
}

func addToDB(db *sql.DB, todo todoitem) error {

	return nil
}
