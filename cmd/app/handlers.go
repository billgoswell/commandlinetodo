package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.viewport.Width = msg.Width
		m.viewport.Height = m.getViewportHeight(msg.Height)
	case tea.KeyMsg:
		// Check list selector mode first
		if m.listSelectorOpen {
			return m.handleListSelector(msg)
		}
		if m.editMode {
			return m.handleEditMode(msg)
		}
		if m.deleteConfirm {
			return m.handleDeleteConfirm(msg)
		}
		if m.inputActive {
			switch m.inputMode {
			case InputModeTask:
				return m.handleTaskInput(msg)
			case InputModePriority:
				return m.handlePriorityInput(msg)
			case InputModeDueDate:
				return m.handleDueDateInput(msg)
			case InputModeListName:
				return m.handleListNameInput(msg)
			}
		}
		return m.handleMainKeyboard(msg)
	}
	return m, nil
}

func (m model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.resetEditState()
		return m, nil
	case KeyEnter:
		editedText := m.textInput.Value()
		if editedText != "" && m.editIndex >= 0 && m.editIndex < len(m.items) {
			m.items[m.editIndex].todo = editedText
			m.items[m.editIndex].todolistID = m.currentListID
			if err := updateItemInDB(m.items[m.editIndex]); err != nil {
				fmt.Println("Error updating task:", err)
			}
		}
		m.resetEditState()
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyY:
		actualIndex := m.getVisibleItemActualIndex(m.confirmDeleteIndex)
		if actualIndex >= 0 && actualIndex < len(m.items) {
			if err := markItemAsDeleted(m.items[actualIndex].id); err != nil {
				fmt.Println("Error deleting task:", err)
			}
			m.items = append(m.items[:actualIndex], m.items[actualIndex+1:]...)
			if m.cursor >= m.getVisibleItemCount() && m.cursor > 0 {
				m.cursor--
			}
			m.sortItems()
		}
		m.resetDeleteConfirm()
		return m, nil
	case KeyN, KeyEsc:
		m.resetDeleteConfirm()
		return m, nil
	}
	return m, nil
}

func (m model) handleTaskInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.resetInputState()
		return m, nil
	case KeyEnter:
		taskText := m.textInput.Value()
		if taskText != "" {
			m.newTaskText = taskText
			m.inputMode = InputModePriority
			m.newPriority = DefaultPriority
			m.viewport.Height = m.getViewportHeight(m.height)
			return m, nil
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m model) handlePriorityInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.inputMode = InputModeTask
		m.textInput.Reset()
		m.newTaskText = ""
		m.newPriority = DefaultPriority
		m.viewport.Height = m.getViewportHeight(m.height)
		return m, nil
	case KeyEnter:
		m.inputMode = InputModeDueDate
		m.textInput.Reset()
		m.newDueDate = 0
		m.viewport.Height = m.getViewportHeight(m.height)
		return m, nil
	case KeyPriority1:
		m.newPriority = PriorityHigh
		return m, nil
	case KeyPriority2:
		m.newPriority = PriorityMedHigh
		return m, nil
	case KeyPriority3:
		m.newPriority = PriorityMed
		return m, nil
	case KeyPriority4:
		m.newPriority = PriorityLow
		return m, nil
	case KeyUp, KeyK:
		if m.newPriority > PriorityHigh {
			m.newPriority--
		}
		return m, nil
	case KeyDown, KeyJ:
		if m.newPriority < PriorityLow {
			m.newPriority++
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleDueDateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		if m.editDueDateMode {
			m.resetInputState()
			return m, nil
		}
		m.inputMode = InputModePriority
		m.textInput.Reset()
		m.newDueDate = 0
		m.viewport.Height = m.getViewportHeight(m.height)
		return m, nil
	case KeyEnter:
		duedate := parseDueDate(m.textInput.Value())
		if m.editDueDateMode {
			// Editing due date on existing task
			if m.editIndex >= 0 && m.editIndex < len(m.items) {
				m.items[m.editIndex].duedate = duedate
				m.items[m.editIndex].todolistID = m.currentListID
				if err := updateItemInDB(m.items[m.editIndex]); err != nil {
					fmt.Println("Error updating task:", err)
				}
			}
			m.resetInputState()
			return m, nil
		}
		// Creating new task
		newTask := todoitem{
			done:       false,
			todo:       m.newTaskText,
			priority:   m.newPriority,
			dateadded:  time.Now().Unix(),
			duedate:    duedate,
			todolistID: m.currentListID,
		}
		if err := saveItemToDB(newTask); err != nil {
			fmt.Println("Error saving task:", err)
		}
		m.items = append(m.items, newTask)
		m.sortItems()
		m.cursor = 0
		m.resetInputState()
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m model) handleMainKeyboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyCtrlC, KeyQ:
		return m, tea.Quit
	case KeyW:
		m.listSelectorOpen = true
		m.listCursorPos = m.currentListIndex
		m.viewport.Height = m.getViewportHeight(m.height)
		return m, nil
	case KeyUp, KeyK:
		if m.cursor > 0 {
			m.cursor--
		}
	case KeyDown, KeyJ:
		visibleCount := m.getVisibleItemCount()
		if m.cursor < visibleCount-1 {
			m.cursor++
		}
	case KeyEnter:
		if m.cursor < m.getVisibleItemCount() {
			m.toggleTaskDone(m.cursor)
		}
	case KeySpace:
		if m.cursor < m.getVisibleItemCount() {
			m.toggleTaskDone(m.cursor)
		}
	case KeyD:
		if m.cursor < m.getVisibleItemCount() {
			m.deleteConfirm = true
			m.confirmDeleteIndex = m.cursor
			m.viewport.Height = m.getViewportHeight(m.height)
		}
	case KeyA:
		m.inputActive = true
		m.inputMode = InputModeTask
		m.textInput.Reset()
		m.textInput.Focus()
		m.viewport.Height = m.getViewportHeight(m.height)
	case KeyE:
		if m.cursor < m.getVisibleItemCount() {
			m.editMode = true
			m.editIndex = m.getVisibleItemActualIndex(m.cursor)
			m.textInput.SetValue(m.items[m.editIndex].todo)
			m.textInput.Focus()
			m.viewport.Height = m.getViewportHeight(m.height)
		}
	case KeyT:
		if m.cursor < m.getVisibleItemCount() {
			m.inputActive = true
			m.inputMode = InputModeDueDate
			m.editIndex = m.getVisibleItemActualIndex(m.cursor)
			m.editDueDateMode = true
			m.textInput.Reset()
			m.textInput.Focus()
			m.viewport.Height = m.getViewportHeight(m.height)
		}
	}
	return m, nil
}

func (m model) toggleTaskDone(visibleIndex int) {
	actualIndex := m.getVisibleItemActualIndex(visibleIndex)
	if actualIndex < 0 || actualIndex >= len(m.items) {
		return
	}
	m.items[actualIndex].done = !m.items[actualIndex].done
	if m.items[actualIndex].done {
		m.items[actualIndex].datecompleted = time.Now().Unix()
	} else {
		m.items[actualIndex].datecompleted = 0
	}
	if err := updateItemInDB(m.items[actualIndex]); err != nil {
		fmt.Println("Error updating task:", err)
	}
	m.sortItems()
}

func (m model) handleListSelector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle delete confirmation
	if m.listDeleteConfirm {
		switch msg.String() {
		case KeyY:
			if m.listCursorPos < len(m.todolists) {
				selectedListID := m.todolists[m.listCursorPos].id
				if err := deleteTodoList(selectedListID); err != nil {
					fmt.Println("Error deleting list:", err)
					m.resetListManageState()
					return m, nil
				}
				// Remove deleted list from todolists
				m.todolists = append(m.todolists[:m.listCursorPos], m.todolists[m.listCursorPos+1:]...)
				if m.listCursorPos >= len(m.todolists) && m.listCursorPos > 0 {
					m.listCursorPos--
				}
				// If deleted list was current, switch to first available
				if selectedListID == m.currentListID && len(m.todolists) > 0 {
					m.currentListIndex = 0
					m.currentListID = m.todolists[0].id
					m.cursor = 0
					m.scrollOffset = 0
				}
			}
			m.resetListManageState()
			return m, nil
		case KeyN, KeyEsc:
			m.listDeleteConfirm = false
			m.listManageMode = true
			return m, nil
		}
		return m, nil
	}

	// Handle manage mode
	if m.listManageMode {
		switch msg.String() {
		case KeyR:
			// Enter rename mode
			if m.listCursorPos < len(m.todolists) {
				m.listManageMode = false
				m.inputActive = true
				m.inputMode = InputModeListName
				m.listManageAction = "rename"
				m.textInput.SetValue(m.todolists[m.listCursorPos].name)
				m.textInput.Focus()
				m.viewport.Height = m.getViewportHeight(m.height)
			}
			return m, nil
		case KeyD:
			// Show delete confirmation
			m.listDeleteConfirm = true
			return m, nil
		case KeyA:
			// Toggle archive
			if m.listCursorPos < len(m.todolists) {
				selectedList := m.todolists[m.listCursorPos]
				if selectedList.archived {
					if err := unarchiveTodoList(selectedList.id); err != nil {
						fmt.Println("Error unarchiving list:", err)
						return m, nil
					}
					m.todolists[m.listCursorPos].archived = false
				} else {
					if err := archiveTodoList(selectedList.id); err != nil {
						fmt.Println("Error archiving list:", err)
						return m, nil
					}
					m.todolists[m.listCursorPos].archived = true
					// If archived list was current, switch to first available non-archived
					if selectedList.id == m.currentListID {
						for i, list := range m.todolists {
							if !list.archived {
								m.currentListIndex = i
								m.currentListID = list.id
								m.cursor = 0
								m.scrollOffset = 0
								break
							}
						}
					}
				}
			}
			m.resetListManageState()
			return m, nil
		case KeyEsc:
			m.listManageMode = false
			return m, nil
		}
		return m, nil
	}

	// Normal list selector navigation
	switch msg.String() {
	case KeyUp, KeyK:
		if m.listCursorPos > 0 {
			m.listCursorPos--
		}
		return m, nil
	case KeyDown, KeyJ:
		if m.listCursorPos < len(m.todolists) {
			m.listCursorPos++
		}
		return m, nil
	case KeyM:
		// Enter manage mode (only if cursor on a list, not on "Create New List")
		if m.listCursorPos < len(m.todolists) {
			m.listManageMode = true
			m.viewport.Height = m.getViewportHeight(m.height)
		}
		return m, nil
	case KeyEnter:
		// Check if "Create New List" is selected
		if m.listCursorPos == len(m.todolists) {
			m.listSelectorOpen = false
			m.inputActive = true
			m.inputMode = InputModeListName
			m.textInput.Reset()
			m.textInput.Focus()
			m.viewport.Height = m.getViewportHeight(m.height)
			return m, nil
		}
		// Switch to selected list
		if m.listCursorPos < len(m.todolists) {
			m.currentListIndex = m.listCursorPos
			m.currentListID = m.todolists[m.listCursorPos].id
			m.cursor = 0
			m.scrollOffset = 0
			// Reload items for the new list
			m.listSelectorOpen = false
			m.viewport.Height = m.getViewportHeight(m.height)
		}
		return m, nil
	case KeyN:
		// Create new list from list selector
		m.listSelectorOpen = false
		m.inputActive = true
		m.inputMode = InputModeListName
		m.textInput.Reset()
		m.textInput.Focus()
		m.viewport.Height = m.getViewportHeight(m.height)
		return m, nil
	case KeyEsc:
		m.listSelectorOpen = false
		m.resetListManageState()
		m.viewport.Height = m.getViewportHeight(m.height)
		return m, nil
	}
	return m, nil
}

func (m model) handleListNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEsc:
		m.resetListNameInput()
		return m, nil
	case KeyEnter:
		listName := m.textInput.Value()
		if listName != "" {
			if m.listManageAction == "rename" {
				// Rename existing list
				if m.listCursorPos < len(m.todolists) {
					listID := m.todolists[m.listCursorPos].id
					if err := updateTodoListName(listID, listName); err != nil {
						fmt.Println("Error renaming list:", err)
						m.resetListNameInput()
						return m, nil
					}
					m.todolists[m.listCursorPos].name = listName
				}
			} else {
				// Create new list
				newID, err := createTodoList(listName)
				if err != nil {
					fmt.Println("Error creating list:", err)
					m.resetListNameInput()
					return m, nil
				}
				// Reload todolists from database
				todolists, err := getTodoLists()
				if err != nil {
					fmt.Println("Error loading todolists:", err)
					m.resetListNameInput()
					return m, nil
				}
				m.todolists = todolists
				// Switch to the new list
				for i, list := range m.todolists {
					if list.id == newID {
						m.currentListIndex = i
						m.currentListID = newID
						break
					}
				}
				m.cursor = 0
				m.scrollOffset = 0
			}
		}
		m.listSelectorOpen = false
		m.resetListNameInput()
		return m, nil
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m *model) resetInputState() {
	m.inputActive = false
	m.inputMode = ""
	m.newTaskText = ""
	m.newPriority = DefaultPriority
	m.newDueDate = 0
	m.editDueDateMode = false
	m.textInput.Reset()
}

func (m *model) resetListNameInput() {
	m.inputActive = false
	m.inputMode = ""
	m.listManageAction = ""
	m.textInput.Reset()
}

func (m *model) resetListManageState() {
	m.listManageMode = false
	m.listManageAction = ""
	m.listDeleteConfirm = false
}

func (m *model) resetEditState() {
	m.editMode = false
	m.editIndex = -1
	m.textInput.Reset()
}

func (m *model) resetDeleteConfirm() {
	m.deleteConfirm = false
	m.confirmDeleteIndex = -1
}
