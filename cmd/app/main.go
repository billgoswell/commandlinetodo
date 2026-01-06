package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := LoadConfig()

	_, err := initDB(cfg.DBPath)
	if err != nil {
		logErrorMsg("initialize database", err)
		os.Exit(1)
	}
	defer db.Close()

	todoLists, err := getTodoLists()
	if err != nil {
		logErrorMsg("load todo lists", err)
		os.Exit(1)
	}

	if len(todoLists) == 0 {
		id, err := createTodoList(DefaultListName)
		if err != nil {
			logErrorMsg("create default todo list", err)
			os.Exit(1)
		}
		todoLists = []todoList{{id: id, name: DefaultListName, archived: false}}
	}

	todoItems, err := getItemsFromDB()
	if err != nil {
		logErrorMsg("load items from database", err)
		os.Exit(1)
	}
	m := initialModel(todoItems, todoLists)
	m.sortItems()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tea: %v", err)
		os.Exit(1)
	}
}
