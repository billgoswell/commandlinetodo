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
		return
	}

	// Load todolists
	todolists, err := getTodoLists()
	if err != nil {
		fmt.Println("Failed to load todolists:", err)
		return
	}

	// If no lists exist, create a default "General" list
	if len(todolists) == 0 {
		id, err := createTodoList("General")
		if err != nil {
			fmt.Println("Failed to create default todolist:", err)
			return
		}
		todolists = []todolist{{id: id, name: "General", archived: false}}
	}

	todoitems, err := getItemsFromDB()
	if err != nil {
		fmt.Println("Failed to load items:", err)
		return
	}
	m := intialModel(todoitems, todolists)
	m.sortItems()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tea: %v", err)
	}
	os.Exit(1)
}
