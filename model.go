package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	items              []todoitem
	cursor             int
	width              int
	height             int
	textInput          textinput.Model
	inputActive        bool
	inputMode          string // "task", "priority", or "duedate"
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
}

func intialModel(todoitems []todoitem) model {
	ti := textinput.New()
	ti.Placeholder = "Enter task description..."
	ti.Focus()
	ti.CharLimit = TextInputCharLimit
	ti.Width = TextInputWidth

	vp := viewport.New(ViewportWidth, ViewportHeight)

	return model{
		items:              todoitems,
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
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) sortItems() {
	// Sort by: done status (uncompleted first), then by priority (1 is highest)
	for i := 0; i < len(m.items); i++ {
		for j := i + 1; j < len(m.items); j++ {
			// Compare done status first (false < true, so uncompleted come first)
			if m.items[i].done && !m.items[j].done {
				m.items[i], m.items[j] = m.items[j], m.items[i]
			} else if m.items[i].done == m.items[j].done {
				// If same done status, sort by priority (lower number = higher priority)
				if m.items[i].priority > m.items[j].priority {
					m.items[i], m.items[j] = m.items[j], m.items[i]
				}
			}
		}
	}
}
