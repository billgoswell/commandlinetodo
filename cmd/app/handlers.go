package main

import (
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
		m.errorMsg = ""
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
			m.items[m.editIndex].todoListID = m.currentListID
			if err := updateItemInDB(m.items[m.editIndex]); err != nil {
				m.errorMsg = "Failed to update task: " + err.Error()
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
				m.errorMsg = "Failed to delete task: " + err.Error()
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
		dueDate := parseDueDate(m.textInput.Value())
		if m.editDueDateMode {
			// Editing due date on existing task
			if m.editIndex >= 0 && m.editIndex < len(m.items) {
				m.items[m.editIndex].dueDate = dueDate
				m.items[m.editIndex].todoListID = m.currentListID
				if err := updateItemInDB(m.items[m.editIndex]); err != nil {
					m.errorMsg = "Failed to update task: " + err.Error()
				}
			}
			m.resetInputState()
			return m, nil
		}
		// Creating new task
		newTask := todoItem{
			done:       false,
			todo:       m.newTaskText,
			priority:   m.newPriority,
			dateAdded:  time.Now().Unix(),
			dueDate:    dueDate,
			todoListID: m.currentListID,
		}
		if err := saveItemToDB(newTask); err != nil {
			m.errorMsg = "Failed to save task: " + err.Error()
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
	case KeyL:
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
		m.items[actualIndex].dateCompleted = time.Now().Unix()
	} else {
		m.items[actualIndex].dateCompleted = 0
	}
	if err := updateItemInDB(m.items[actualIndex]); err != nil {
		m.errorMsg = "Failed to update task: " + err.Error()
	}
	m.sortItems()
}

func (m model) handleListSelector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.listDeleteConfirm {
		return m.handleListDeleteConfirm(msg)
	}
	if m.listManageMode {
		return m.handleListManageMode(msg)
	}
	return m.handleListNavigation(msg)
}

func (m model) handleListDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyY:
		m.deleteCurrentList()
	case KeyN, KeyEsc:
		m.listDeleteConfirm = false
		m.listManageMode = true
	}
	return m, nil
}

func (m model) handleListManageMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyR:
		m.enterRenameMode()
	case KeyD:
		m.listDeleteConfirm = true
	case KeyA:
		m.toggleListArchive()
	case KeyEsc:
		m.listManageMode = false
	}
	return m, nil
}

func (m model) handleListNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyUp, KeyK:
		if m.listCursorPos > 0 {
			m.listCursorPos--
		}
	case KeyDown, KeyJ:
		if m.listCursorPos < len(m.todoLists)-1 {
			m.listCursorPos++
		}
	case KeyM:
		if m.listCursorPos < len(m.todoLists) {
			m.listManageMode = true
			m.viewport.Height = m.getViewportHeight(m.height)
		}
	case KeyEnter:
		m.handleListSelection()
	case KeyN:
		m.startCreateNewList()
	case KeyEsc:
		m.closeListSelector()
	}
	return m, nil
}

// Helper functions for list operations
func (m *model) deleteCurrentList() {
	if m.listCursorPos >= len(m.todoLists) {
		m.resetListManageState()
		return
	}

	selectedListID := m.todoLists[m.listCursorPos].id
	if err := deleteTodoList(selectedListID); err != nil {
		m.errorMsg = "Failed to delete list: " + err.Error()
		m.resetListManageState()
		return
	}

	// Remove tasks belonging to this list from memory
	var remainingItems []todoItem
	for _, item := range m.items {
		if item.todoListID != selectedListID {
			remainingItems = append(remainingItems, item)
		}
	}
	m.items = remainingItems
	m.invalidateCache()

	m.todoLists = append(m.todoLists[:m.listCursorPos], m.todoLists[m.listCursorPos+1:]...)
	if m.listCursorPos >= len(m.todoLists) && m.listCursorPos > 0 {
		m.listCursorPos--
	}

	if selectedListID == m.currentListID && len(m.todoLists) > 0 {
		m.switchToList(0)
	}
	m.resetListManageState()
}

func (m *model) enterRenameMode() {
	if m.listCursorPos >= len(m.todoLists) {
		return
	}
	m.listManageMode = false
	m.inputActive = true
	m.inputMode = InputModeListName
	m.listManageAction = "rename"
	m.textInput.SetValue(m.todoLists[m.listCursorPos].name)
	m.textInput.Focus()
	m.viewport.Height = m.getViewportHeight(m.height)
}

func (m *model) toggleListArchive() {
	if m.listCursorPos >= len(m.todoLists) {
		return
	}

	selectedList := m.todoLists[m.listCursorPos]
	if selectedList.archived {
		if err := unarchiveTodoList(selectedList.id); err != nil {
			m.errorMsg = "Failed to unarchive list: " + err.Error()
			return
		}
		m.todoLists[m.listCursorPos].archived = false
	} else {
		if err := archiveTodoList(selectedList.id); err != nil {
			m.errorMsg = "Failed to archive list: " + err.Error()
			return
		}
		m.todoLists[m.listCursorPos].archived = true
		if selectedList.id == m.currentListID {
			m.switchToFirstNonArchivedList()
		}
	}
	m.resetListManageState()
}

func (m *model) switchToFirstNonArchivedList() {
	for i, list := range m.todoLists {
		if !list.archived {
			m.switchToList(i)
			return
		}
	}
}

func (m *model) handleListSelection() {
	// "Create New List" option
	if m.listCursorPos == len(m.todoLists) {
		m.startCreateNewList()
		return
	}
	// Switch to selected list
	if m.listCursorPos < len(m.todoLists) {
		m.switchToList(m.listCursorPos)
		m.closeListSelector()
	}
}

func (m *model) switchToList(index int) {
	if index < 0 || index >= len(m.todoLists) {
		return
	}
	m.currentListIndex = index
	m.currentListID = m.todoLists[index].id
	m.cursor = 0
	m.scrollOffset = 0
	m.invalidateCache()
}

func (m *model) startCreateNewList() {
	m.listSelectorOpen = false
	m.inputActive = true
	m.inputMode = InputModeListName
	m.textInput.Reset()
	m.textInput.Focus()
	m.viewport.Height = m.getViewportHeight(m.height)
}

func (m *model) closeListSelector() {
	m.listSelectorOpen = false
	m.resetListManageState()
	m.viewport.Height = m.getViewportHeight(m.height)
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
				if m.listCursorPos < len(m.todoLists) {
					listID := m.todoLists[m.listCursorPos].id
					if err := updateTodoListName(listID, listName); err != nil {
						m.errorMsg = "Failed to rename list: " + err.Error()
						m.resetListNameInput()
						return m, nil
					}
					m.todoLists[m.listCursorPos].name = listName
				}
			} else {
				// Create new list
				newID, err := createTodoList(listName)
				if err != nil {
					m.errorMsg = "Failed to create list: " + err.Error()
					m.resetListNameInput()
					return m, nil
				}
				// Reload todolists from database
				todolists, err := getTodoLists()
				if err != nil {
					m.errorMsg = "Failed to load lists: " + err.Error()
					m.resetListNameInput()
					return m, nil
				}
				m.todoLists = todolists
				// Switch to the new list
				for i, list := range m.todoLists {
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

// Unified state reset function
func (m *model) resetAllState() {
	m.inputActive = false
	m.inputMode = ""
	m.newTaskText = ""
	m.newPriority = DefaultPriority
	m.newDueDate = 0
	m.editDueDateMode = false
	m.editMode = false
	m.editIndex = -1
	m.deleteConfirm = false
	m.confirmDeleteIndex = -1
	m.listManageMode = false
	m.listManageAction = ""
	m.listDeleteConfirm = false
	m.textInput.Reset()
}

// Specific reset functions for convenience (delegate to resetAllState where possible)
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
