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
	if m.errorMsg != "" {
		s = append(s, ErrorStyle.Render("Error: "+m.errorMsg))
	}
	s = append(s, "Press l for lists, a to add, e to edit, t to set due date, d to delete, q to quit.\n")
	return strings.Join(s, "\n")
}

func (m *model) priorityDisplay() string {
	lines := []string{"Select priority:"}
	priorities := []int{PriorityHigh, PriorityMedHigh, PriorityMed, PriorityLow}

	for _, priority := range priorities {
		label := fmt.Sprintf("%d: %s", priority, PriorityLabels[priority])
		styledLabel := PriorityStyles[priority].Render(label)

		if m.newPriority == priority {
			lines = append(lines, SelectedStyle.Render("▶ "+styledLabel))
		} else {
			lines = append(lines, "  "+styledLabel)
		}
	}
	return strings.Join(lines, "\n")
}

func (m *model) updateViewport() {
	// Filter items for the current list
	visibleItems := m.filterItemsByList(m.currentListID)
	viewportHeight := m.viewport.Height

	// Adjust cursor to stay within bounds
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

func (m *model) formatTaskTimestamps(item todoItem) string {
	var dateStr string
	if item.dateAdded > 0 {
		t := time.Unix(item.dateAdded, 0)
		dateStr = "added " + timediff.TimeDiff(t)
	}

	// Add completion time if task is done
	if item.done && item.dateCompleted > 0 {
		t := time.Unix(item.dateCompleted, 0)
		completedStr := "completed " + timediff.TimeDiff(t)
		dateStr = dateStr + " | " + completedStr
	}

	return dateStr
}

func (m *model) formatDueDate(item todoItem) string {

	if item.dueDate == 0 {
		return ""
	}

	now := time.Now().Unix()
	if item.dueDate < now {
		return " | OVERDUE"
	}

	// Calculate days until due
	daysUntil := (item.dueDate - now) / SecondsPerDay
	if daysUntil == 0 {
		return " | due today"
	}

	dueStr := fmt.Sprintf(" | due in %d day", daysUntil)
	if daysUntil > 1 {
		dueStr += "s"
	}
	return dueStr

}

func (m model) getStyle(i int, item todoItem) lipgloss.Style {
	if m.cursor == i && item.done {
		return SelectedDoneStyle
	}

	isOverdue := item.dueDate > 0 && item.dueDate < time.Now().Unix() && !item.done
	if isOverdue {
		if m.cursor == i {
			return OverdueStyle
		}
		return OverdueStyle
	}

	if m.cursor == i {
		if selectedStyle, exists := SelectedPriorityStyles[item.priority]; exists {
			return selectedStyle
		}
		return SelectedStyle
	}

	if item.done {
		return DoneStyle
	}

	if style, exists := PriorityStyles[item.priority]; exists {
		return style
	}
	return ItemStyle
}

func (m model) getViewportHeight(totalHeight int) int {
	availableHeight := totalHeight - ViewportBaseLines
	additionalHeight := 0

	if m.listSelectorOpen {
		// List selector height = header + lists + "Create New List" option + help text + spacing
		additionalHeight = len(m.todoLists) * HeightAdjustListPerItem
		additionalHeight += 5 // Header, spacing, help text
	} else if m.editMode || m.inputActive {
		additionalHeight = HeightAdjustDefault
		if m.inputMode == InputModePriority {
			additionalHeight = HeightAdjustPriority
		}
	} else if m.deleteConfirm {
		additionalHeight = HeightAdjustDelete
	}

	return availableHeight - additionalHeight
}

func (m *model) getCurrentListName() string {
	if m.currentListIndex >= 0 && m.currentListIndex < len(m.todoLists) {
		return m.todoLists[m.currentListIndex].name
	}
	return "Unknown"
}

func (m *model) renderListSelector() string {
	var lines []string

	// Show delete confirmation if needed
	if m.listDeleteConfirm && m.listCursorPos < len(m.todoLists) {
		selectedList := m.todoLists[m.listCursorPos]
		taskCount := m.countTasksInList(selectedList.id)
		lines = append(lines, TitleStyle.Render("Delete List"))
		lines = append(lines, fmt.Sprintf("Delete '%s'? This will delete %d task(s).", selectedList.name, taskCount))
		lines = append(lines, SelectedStyle.Render("Confirm: (y/n)"))
		return strings.Join(lines, "\n")
	}

	// Show management options if in manage mode
	if m.listManageMode && m.listCursorPos < len(m.todoLists) {
		lines = append(lines, TitleStyle.Render("Manage List"))
		lines = append(lines, fmt.Sprintf("List: %s", m.todoLists[m.listCursorPos].name))
		lines = append(lines, "")
		lines = append(lines, "r: Rename")
		lines = append(lines, "d: Delete")
		if m.todoLists[m.listCursorPos].archived {
			lines = append(lines, "a: Unarchive")
		} else {
			lines = append(lines, "a: Archive")
		}
		lines = append(lines, TitleStyle.Render("(Press key or Esc to go back)"))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, TitleStyle.Render("Select list:"))

	// Render list options
	for i, list := range m.todoLists {
		if i == m.listCursorPos {
			lines = append(lines, SelectedStyle.Render("▶ "+list.name))
		} else {
			lines = append(lines, "  "+list.name)
		}
	}

	// Add create new list option
	lines = append(lines, "")
	if m.listCursorPos == len(m.todoLists) {
		lines = append(lines, SelectedStyle.Render("▶ Create New List (n)"))
	} else {
		lines = append(lines, "  Create New List (n)")
	}

	// Show hint for manage mode
	var hint string
	if m.listCursorPos < len(m.todoLists) {
		hint = "(Use k/↑ and j/↓ to navigate, Enter to select, m for manage, Esc to cancel)"
	} else {
		hint = "(Use k/↑ and j/↓ to navigate, Enter to select, Esc to cancel)"
	}
	lines = append(lines, TitleStyle.Render(hint))

	return strings.Join(lines, "\n")
}

// filterItemsByList returns all items for a specific list (cached)
func (m *model) filterItemsByList(listID int) []todoItem {
	if m.filteredListID == listID && m.filteredItems != nil {
		return m.filteredItems
	}
	var filtered []todoItem
	for _, item := range m.items {
		if item.todoListID == listID {
			filtered = append(filtered, item)
		}
	}
	m.filteredItems = filtered
	m.filteredListID = listID
	return filtered
}

func (m *model) invalidateCache() {
	m.filteredItems = nil
}

func (m *model) getVisibleItemCount() int {
	return len(m.filterItemsByList(m.currentListID))
}

func (m *model) getVisibleItemActualIndex(visibleIndex int) int {
	filteredItems := m.filterItemsByList(m.currentListID)
	if visibleIndex < 0 || visibleIndex >= len(filteredItems) {
		return -1
	}
	// Find the actual index in m.items
	targetItem := filteredItems[visibleIndex]
	for i, item := range m.items {
		if item.id == targetItem.id {
			return i
		}
	}
	return -1
}

func (m *model) countTasksInList(listID int) int {
	return len(m.filterItemsByList(listID))
}
