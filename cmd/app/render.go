package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mergestat/timediff"
)

func (m model) View() string {
	currentListName := m.getCurrentListName()
	s := []string{TitleStyle.Render("Todo list: " + currentListName)}

	// Update viewport content
	m.updateViewport()
	s = append(s, m.viewport.View())

	// Render list selector modal if open
	if m.listSelectorOpen {
		s = append(s, "")
		s = append(s, m.renderListSelector())
	} else if m.editMode {
		s = append(s, "")
		s = append(s, TitleStyle.Render("Edit task:"))
		s = append(s, m.textInput.View())
		s = append(s, TitleStyle.Render("(Press Enter to save, Esc to cancel)"))
	} else if m.deleteConfirm {
		s = append(s, "")
		s = append(s, SelectedStyle.Render("Delete this task? (y/n)"))
	} else if m.inputActive {
		switch m.inputMode {
		case InputModeTask:
			s = append(s, TitleStyle.Render("New task:"))
			s = append(s, m.textInput.View())
			s = append(s, TitleStyle.Render("(Press Enter to continue, Esc to cancel)"))
		case InputModePriority:
			s = append(s, TitleStyle.Render("Task: "+m.newTaskText))
			s = append(s, "")
			s = append(s, m.priorityDisplay())
			s = append(s, TitleStyle.Render("(Use k/↑ and j/↓ to navigate, 1-4 to jump, Enter to save, Esc to go back)"))
		case InputModeDueDate:
			if m.editDueDateMode {
				s = append(s, TitleStyle.Render("Edit due date:"))
			} else {
				s = append(s, TitleStyle.Render("Due date (optional):"))
			}
			s = append(s, m.textInput.View())
			s = append(s, TitleStyle.Render("(Enter days like '3' or date like '12/25/2025', press Enter to skip, Esc to cancel)"))
		}
	}
	s = append(s, "Press w for lists, a to add, e to edit, t to set due date, d to delete, q to quit.\n")
	return strings.Join(s, "\n")
}

func (m *model) priorityDisplay() string {
	priorityDisplay := "Select priority:\n"
	priorityColors := []string{
		OneStyle.Render("1: 🟥 High"),
		TwoStyle.Render("2: 🟧 Medium-High"),
		ThreeStyle.Render("3: 🟨 Medium"),
		FourStyle.Render("4: 🟩 Low"),
	}
	for i, p := range priorityColors {
		if m.newPriority == i+1 {
			priorityDisplay += SelectedStyle.Render("▶ "+p) + "\n"
		} else {
			priorityDisplay += "  " + p + "\n"
		}
	}
	return priorityDisplay
}

func (m *model) updateViewport() {
	// Calculate how many lines each task takes
	viewportHeight := m.viewport.Height

	// Filter items for the current list
	var visibleItems []todoitem
	for _, item := range m.items {
		if item.todolistID == m.currentListID {
			visibleItems = append(visibleItems, item)
		}
	}

	// Adjust cursor to stay within bounds of visible items
	if m.cursor >= len(visibleItems) {
		m.cursor = len(visibleItems) - 1
	}
	if m.cursor < 0 && len(visibleItems) > 0 {
		m.cursor = 0
	}

	// Adjust scroll offset to keep cursor visible
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	tasksPerView := viewportHeight / LinesPerTask
	if m.cursor >= m.scrollOffset+tasksPerView {
		m.scrollOffset = m.cursor - tasksPerView + 1
	}

	var taskLines []string

	for i, item := range visibleItems {
		dateStr := m.formatTaskTimestamps(item)
		dueStr := m.formatDueDate(item)

		c := fmt.Sprintf("%s\n%s%s\n", item.todo, dateStr, dueStr)

		style := m.getStyle(i, item)
		c = style.Render(c)
		taskLines = append(taskLines, c)
	}

	m.viewport.SetContent(strings.Join(taskLines, "\n"))

	m.viewport.YOffset = m.scrollOffset * LinesPerTask
}

func (m *model) formatTaskTimestamps(item todoitem) string {
	var dateStr string
	if item.dateadded > 0 {
		t := time.Unix(item.dateadded, 0)
		dateStr = "added " + timediff.TimeDiff(t)
	}

	// Add completion time if task is done
	if item.done && item.datecompleted > 0 {
		t := time.Unix(item.datecompleted, 0)
		completedStr := "completed " + timediff.TimeDiff(t)
		dateStr = dateStr + " | " + completedStr
	}

	return dateStr
}

func (m *model) formatDueDate(item todoitem) string {

	if item.duedate == 0 {
		return ""
	}

	now := time.Now().Unix()
	if item.duedate < now {
		return " | OVERDUE"
	}

	// Calculate days until due
	daysUntil := (item.duedate - now) / SecondsPerDay
	if daysUntil == 0 {
		return " | due today"
	}

	dueStr := fmt.Sprintf(" | due in %d day", daysUntil)
	if daysUntil > 1 {
		dueStr += "s"
	}
	return dueStr

}

func (m model) getStyle(i int, item todoitem) lipgloss.Style {
	var style lipgloss.Style
	if m.cursor == i && item.done {
		style = SelectedDoneStyle
	} else if m.cursor == i {
		// Check if overdue
		if item.duedate > 0 && item.duedate < time.Now().Unix() && !item.done {
			style = OverdueStyle
		} else {
			switch item.priority {
			case 1:
				style = SelectedOneStyle
			case 2:
				style = SelectedTwoStyle
			case 3:
				style = SelectedThreeStyle
			case 4:
				style = SelectedFourStyle
			default:
				style = SelectedStyle
			}
		}
	} else if item.done {
		style = DoneStyle
	} else {
		// Check if overdue
		if item.duedate > 0 && item.duedate < time.Now().Unix() {
			style = OverdueStyle
		} else {
			switch item.priority {
			case 1:
				style = OneStyle
			case 2:
				style = TwoStyle
			case 3:
				style = ThreeStyle
			case 4:
				style = FourStyle
			}
		}
	}
	return style
}

func (m model) getViewportHeight(totalHeight int) int {
	availableHeight := totalHeight - ViewportBaseLines
	additionalHeight := 0
	if m.listSelectorOpen {
		additionalHeight = len(m.todolists) + 5
	} else if m.editMode || m.inputActive {
		additionalHeight = 6
		if m.inputMode == InputModePriority {
			additionalHeight = 11
		}
	} else if m.deleteConfirm {
		additionalHeight = 2
	}

	return availableHeight - additionalHeight
}

func (m *model) getCurrentListName() string {
	if m.currentListIndex >= 0 && m.currentListIndex < len(m.todolists) {
		return m.todolists[m.currentListIndex].name
	}
	return "Unknown"
}

func (m *model) renderListSelector() string {
	var lines []string

	// Show delete confirmation if needed
	if m.listDeleteConfirm && m.listCursorPos < len(m.todolists) {
		selectedList := m.todolists[m.listCursorPos]
		taskCount := m.countTasksInList(selectedList.id)
		lines = append(lines, TitleStyle.Render("Delete List"))
		lines = append(lines, fmt.Sprintf("Delete '%s'? This will delete %d task(s).", selectedList.name, taskCount))
		lines = append(lines, SelectedStyle.Render("Confirm: (y/n)"))
		return strings.Join(lines, "\n")
	}

	// Show management options if in manage mode
	if m.listManageMode && m.listCursorPos < len(m.todolists) {
		lines = append(lines, TitleStyle.Render("Manage List"))
		lines = append(lines, fmt.Sprintf("List: %s", m.todolists[m.listCursorPos].name))
		lines = append(lines, "")
		lines = append(lines, "r: Rename")
		lines = append(lines, "d: Delete")
		if m.todolists[m.listCursorPos].archived {
			lines = append(lines, "a: Unarchive")
		} else {
			lines = append(lines, "a: Archive")
		}
		lines = append(lines, TitleStyle.Render("(Press key or Esc to go back)"))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, TitleStyle.Render("Select list:"))

	// Render list options
	for i, list := range m.todolists {
		if i == m.listCursorPos {
			lines = append(lines, SelectedStyle.Render("▶ "+list.name))
		} else {
			lines = append(lines, "  "+list.name)
		}
	}

	// Add create new list option
	lines = append(lines, "")
	if m.listCursorPos == len(m.todolists) {
		lines = append(lines, SelectedStyle.Render("▶ Create New List (n)"))
	} else {
		lines = append(lines, "  Create New List (n)")
	}

	// Show hint for manage mode
	var hint string
	if m.listCursorPos < len(m.todolists) {
		hint = "(Use k/↑ and j/↓ to navigate, Enter to select, m for manage, Esc to cancel)"
	} else {
		hint = "(Use k/↑ and j/↓ to navigate, Enter to select, Esc to cancel)"
	}
	lines = append(lines, TitleStyle.Render(hint))

	return strings.Join(lines, "\n")
}

func (m *model) getVisibleItemCount() int {
	count := 0
	for _, item := range m.items {
		if item.todolistID == m.currentListID {
			count++
		}
	}
	return count
}

func (m *model) getVisibleItemActualIndex(visibleIndex int) int {
	count := 0
	for i, item := range m.items {
		if item.todolistID == m.currentListID {
			if count == visibleIndex {
				return i
			}
			count++
		}
	}
	return -1
}

func (m *model) countTasksInList(listID int) int {
	count := 0
	for _, item := range m.items {
		if item.todolistID == listID {
			count++
		}
	}
	return count
}
