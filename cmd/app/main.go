package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	_, err := initDB()
	if err != nil {
		fmt.Println("Failed to initialize database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Load todoLists
	todoLists, err := getTodoLists()
	if err != nil {
		fmt.Println("Failed to load todoLists:", err)
		return
	}

	// If no lists exist, create a default "General" list
	if len(todoLists) == 0 {
		id, err := createTodoList("General")
		if err != nil {
			fmt.Println("Failed to create default todoList:", err)
			return
		}
		todoLists = []todoList{{id: id, name: "General", archived: false}}
	}

	todoItems, err := getItemsFromDB()
	if err != nil {
		fmt.Println("Failed to load items:", err)
		return
	}
	m := initialModel(todoItems, todoLists)
	m.sortItems()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tea: %v", err)
		os.Exit(1)
	}
}
