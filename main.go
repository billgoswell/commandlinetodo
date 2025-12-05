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

	todoitems, err := getItemsFromDB()
	if err != nil {
		fmt.Println("Failed to load items:", err)
		return
	}
	m := intialModel(todoitems)
	m.sortItems()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tea: %v", err)
	}
	os.Exit(1)
}
