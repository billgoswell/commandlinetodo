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
		if m.confirmDeleteIndex >= 0 && m.confirmDeleteIndex < len(m.items) {
			if err := markItemAsDeleted(m.items[m.confirmDeleteIndex].id); err != nil {
				fmt.Println("Error deleting task:", err)
			}
			m.items = append(m.items[:m.confirmDeleteIndex], m.items[m.confirmDeleteIndex+1:]...)
			if m.cursor >= len(m.items) && m.cursor > 0 {
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
				if err := updateItemInDB(m.items[m.editIndex]); err != nil {
					fmt.Println("Error updating task:", err)
				}
			}
			m.resetInputState()
			return m, nil
		}
		// Creating new task
		newTask := todoitem{
			done:      false,
			todo:      m.newTaskText,
			priority:  m.newPriority,
			dateadded: time.Now().Unix(),
			duedate:   duedate,
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
	case KeyUp, KeyK:
		if m.cursor > 0 {
			m.cursor--
		}
	case KeyDown, KeyJ:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case KeyEnter:
		m.toggleTaskDone(m.cursor)
	case KeySpace:
		if m.cursor < len(m.items) {
			m.toggleTaskDone(m.cursor)
		}
	case KeyD:
		if m.cursor < len(m.items) {
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
		if m.cursor < len(m.items) {
			m.editMode = true
			m.editIndex = m.cursor
			m.textInput.SetValue(m.items[m.cursor].todo)
			m.textInput.Focus()
			m.viewport.Height = m.getViewportHeight(m.height)
		}
	case KeyT:
		if m.cursor < len(m.items) {
			m.inputActive = true
			m.inputMode = InputModeDueDate
			m.editIndex = m.cursor
			m.editDueDateMode = true
			m.textInput.Reset()
			m.textInput.Focus()
			m.viewport.Height = m.getViewportHeight(m.height)
		}
	}
	return m, nil
}

func (m model) toggleTaskDone(index int) {
	m.items[index].done = !m.items[index].done
	if m.items[index].done {
		m.items[index].datecompleted = time.Now().Unix()
	} else {
		m.items[index].datecompleted = 0
	}
	if err := updateItemInDB(m.items[index]); err != nil {
		fmt.Println("Error updating task:", err)
	}
	m.sortItems()
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

func (m *model) resetEditState() {
	m.editMode = false
	m.editIndex = -1
	m.textInput.Reset()
}

func (m *model) resetDeleteConfirm() {
	m.deleteConfirm = false
	m.confirmDeleteIndex = -1
}
