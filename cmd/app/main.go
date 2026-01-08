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

	// Create data store (local or with sync)
	var store DataStore
	var syncStore *SyncStore
	localStore := NewLocalStore(db)

	if cfg.Sync.Enabled {
		// Create sync client
		syncClient := NewSyncClient(cfg.Sync)

		// Create sync store
		syncStore = NewSyncStore(localStore, syncClient, cfg.Sync)

		// Perform initial sync if online
		if syncClient.IsOnline() {
			if err := syncStore.FullSync(); err != nil {
				fmt.Printf("Warning: initial sync failed: %v\n", err)
			}
		}

		// Start background sync
		syncStore.StartBackgroundSync()

		store = syncStore
	} else {
		store = localStore
	}

	// Load data using the store interface
	todoLists, err := store.GetTodoLists()
	if err != nil {
		logErrorMsg("load todo lists", err)
		os.Exit(1)
	}

	if len(todoLists) == 0 {
		id, err := store.CreateTodoList(DefaultListName)
		if err != nil {
			logErrorMsg("create default todo list", err)
			os.Exit(1)
		}
		todoLists = []todoList{{id: id, name: DefaultListName, archived: false}}
	}

	todoItems, err := store.GetItems()
	if err != nil {
		logErrorMsg("load items from database", err)
		os.Exit(1)
	}

	// Create model with store
	m := initialModel(todoItems, todoLists)
	m.store = store
	m.syncEnabled = cfg.Sync.Enabled
	if cfg.Sync.Enabled && syncStore != nil {
		m.syncStatus.online = syncStore.client.IsOnline()
		m.syncStatus.lastSyncTime = 0 // Will be set during first sync
	}
	m.sortItems()

	// Run the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())
	defer func() {
		if syncStore != nil {
			syncStore.StopBackgroundSync()
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tea: %v", err)
		os.Exit(1)
	}
}
