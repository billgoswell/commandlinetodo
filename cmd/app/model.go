package main

import (
	"sort"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type todoList struct {
	id           int
	name         string
	displayOrder int
	archived     bool
	createdAt    int64
	updatedAt    int64
}

type model struct {
	items              []todoItem
	cursor             int
	width              int
	height             int
	textInput          textinput.Model
	inputActive        bool
	inputMode          string // "task", "priority", "duedate", or "listname"
	newTaskText        string
	newPriority        int
	newDueDate         int64
	deleteConfirm      bool
	confirmDeleteIndex int
	viewport           viewport.Model
	scrollOffset       int
	editMode           bool
	editIndex          int
	editDueDateMode    bool

	// Todolist management
	todoLists          []todoList // All available lists
	currentListID      int        // Currently active list
	currentListIndex   int        // Index in todoLists slice
	listSelectorOpen   bool       // Is list selector modal open?
	listCursorPos      int        // Cursor in list selector
	newListInputActive bool       // Creating new list?
	listManageMode     bool
	listManageAction   string
	listDeleteConfirm  bool
	errorMsg           string
	filteredItems      []todoItem
	filteredListID     int
}

func initialModel(todoItems []todoItem, todoLists []todoList) model {
	ti := textinput.New()
	ti.Placeholder = "Enter task description..."
	ti.Focus()
	ti.CharLimit = TextInputCharLimit
	ti.Width = TextInputWidth

	vp := viewport.New(ViewportWidth, ViewportHeight)

	// Default to first list if available
	currentListID := 0
	currentListIndex := 0
	if len(todoLists) > 0 {
		currentListID = todoLists[0].id
		currentListIndex = 0
	}

	return model{
		items:              todoItems,
		textInput:          ti,
		inputActive:        false,
		inputMode:          "",
		newTaskText:        "",
		newPriority:        DefaultPriority,
		newDueDate:         0,
		deleteConfirm:      false,
		confirmDeleteIndex: -1,
		viewport:           vp,
		scrollOffset:       0,
		editMode:           false,
		editIndex:          -1,
		editDueDateMode:    false,
		todoLists:          todoLists,
		currentListID:      currentListID,
		currentListIndex:   currentListIndex,
		listSelectorOpen:   false,
		listCursorPos:      0,
		newListInputActive: false,
		listManageMode:     false,
		listManageAction:   "",
		listDeleteConfirm:  false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) sortItems() {
	sort.Slice(m.items, func(i, j int) bool {
		if m.items[i].done != m.items[j].done {
			return !m.items[i].done
		}
		return m.items[i].priority < m.items[j].priority
	})
	m.invalidateCache()
}
