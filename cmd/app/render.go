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

	m.updateViewport()
	s = append(s, m.viewport.View())

	switch m.currentState {
	case StateListSelector:
		s = append(s, "")
		s = append(s, m.renderListSelector())
	case StateEditTask:
		s = append(s, "")
		s = append(s, TitleStyle.Render("Edit task:"))
		s = append(s, m.textInput.View())
		s = append(s, TitleStyle.Render("(Press Enter to save, Esc to cancel)"))
	case StateDeleteConfirm:
		s = append(s, "")
		s = append(s, SelectedStyle.Render("Delete this task? (y/n)"))
	case StateTaskInput:
		s = append(s, TitleStyle.Render("New task:"))
		s = append(s, m.textInput.View())
		s = append(s, TitleStyle.Render("(Press Enter to continue, Esc to cancel)"))
	case StatePrioritySelection:
		s = append(s, TitleStyle.Render("Task: "+m.taskFlow.text))
		s = append(s, "")
		s = append(s, m.priorityDisplay())
		s = append(s, TitleStyle.Render("(Use k/↑ and j/↓ to navigate, 1-4 to jump, Enter to save, Esc to go back)"))
	case StateDueDateInput:
		if m.currentSubState == SubStateEditDueDate {
			s = append(s, TitleStyle.Render("Edit due date:"))
		} else {
			s = append(s, TitleStyle.Render("Due date (optional):"))
		}
		s = append(s, m.textInput.View())
		s = append(s, TitleStyle.Render("(Enter days like '3' or date like '12/25/2025', press Enter to skip, Esc to cancel)"))
	case StateListNameInput:
		if m.currentSubState == SubStateListRename {
			s = append(s, TitleStyle.Render("Rename list:"))
		} else {
			s = append(s, TitleStyle.Render("New list name:"))
		}
		s = append(s, m.textInput.View())
		s = append(s, TitleStyle.Render("(Press Enter to save, Esc to cancel)"))
	}
	if m.errorMsg != "" {
		s = append(s, ErrorStyle.Render("Error: "+m.errorMsg))
	}

	// Add sync status if enabled
	if m.syncEnabled {
		s = append(s, m.renderSyncStatus())
	}

	s = append(s, "Press l for lists, a to add, e to edit, t to set due date, d to delete, q to quit.")
	if m.syncEnabled {
		s = append(s, "Press s to sync.")
	}
	s = append(s, "")
	return strings.Join(s, "\n")
}

func (m *model) priorityDisplay() string {
	lines := []string{"Select priority:"}
	priorities := []int{PriorityHigh, PriorityMedHigh, PriorityMed, PriorityLow}

	for _, priority := range priorities {
		label := fmt.Sprintf("%d: %s", priority, PriorityLabels[priority])
		styledLabel := PriorityStyles[priority].Render(label)

		if m.taskFlow.priority == priority {
			lines = append(lines, SelectedStyle.Render("▶ "+styledLabel))
		} else {
			lines = append(lines, "  "+styledLabel)
		}
	}
	return strings.Join(lines, "\n")
}

func (m *model) updateViewport() {
	visibleItems := m.filterItemsByList(m.currentListID)
	viewportHeight := m.viewport.Height

	if m.cursor >= len(visibleItems) {
		m.cursor = len(visibleItems) - 1
	}
	if m.cursor < 0 && len(visibleItems) > 0 {
		m.cursor = 0
	}

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	tasksPerView := viewportHeight / LinesPerTask
	if m.cursor >= m.scrollOffset+tasksPerView {
		m.scrollOffset = m.cursor - tasksPerView + 1
	}

	currentTime := time.Now().Unix()

	var taskLines []string
	for i, item := range visibleItems {
		dateStr := m.formatTaskTimestamps(item)
		dueStr := m.formatDueDate(item)
		c := fmt.Sprintf("%s\n%s%s\n", item.todo, dateStr, dueStr)
		style := m.getStyle(i, item, currentTime)
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

func (m model) getStyle(i int, item todoItem, currentTime int64) lipgloss.Style {
	if m.cursor == i && item.done {
		return SelectedDoneStyle
	}

	if item.dueDate > 0 && item.dueDate < currentTime && !item.done {
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
	additionalHeight := StateHeightAdjustments[m.currentState]

	if m.currentState == StateListSelector {
		additionalHeight = len(m.todoLists) + 5 // 1 line per list + 5 for header/spacing/help
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

	if m.currentSubState == SubStateListDeleteConfirm {
		selectedList := m.getListAtIndex(m.input.listIndex)
		if selectedList != nil {
			taskCount := m.countTasksInList(selectedList.id)
			lines = append(lines, TitleStyle.Render("Delete List"))
			lines = append(lines, fmt.Sprintf("Delete '%s'? This will delete %d task(s).", selectedList.name, taskCount))
			lines = append(lines, SelectedStyle.Render("Confirm: (y/n)"))
			return strings.Join(lines, "\n")
		}
	}

	if m.currentSubState == SubStateListManage {
		selectedList := m.getListAtIndex(m.input.listIndex)
		if selectedList != nil {
			lines = append(lines, TitleStyle.Render("Manage List"))
			lines = append(lines, fmt.Sprintf("List: %s", selectedList.name))
			lines = append(lines, "")
			lines = append(lines, "r: Rename")
			lines = append(lines, "d: Delete")
			if selectedList.archived {
				lines = append(lines, "a: Unarchive")
			} else {
				lines = append(lines, "a: Archive")
			}
			lines = append(lines, TitleStyle.Render("(Press key or Esc to go back)"))
			return strings.Join(lines, "\n")
		}
	}

	lines = append(lines, TitleStyle.Render("Select list:"))

	for i, list := range m.todoLists {
		if i == m.input.listIndex {
			lines = append(lines, SelectedStyle.Render("▶ "+list.name))
		} else {
			lines = append(lines, "  "+list.name)
		}
	}

	lines = append(lines, "")
	if m.input.listIndex == len(m.todoLists) {
		lines = append(lines, SelectedStyle.Render("▶ Create New List (n)"))
	} else {
		lines = append(lines, "  Create New List (n)")
	}

	var hint string
	if m.input.listIndex < len(m.todoLists) {
		hint = "(Use k/↑ and j/↓ to navigate, Enter to select, m for manage, Esc to cancel)"
	} else {
		hint = "(Use k/↑ and j/↓ to navigate, Enter to select, Esc to cancel)"
	}
	lines = append(lines, TitleStyle.Render(hint))

	return strings.Join(lines, "\n")
}

func (m *model) filterItemsByList(listID int) []todoItem {
	if m.filteredListID == listID && m.cacheValid {
		return m.filteredItems
	}
	var filtered []todoItem
	var indices []int
	for i, item := range m.items {
		if item.todoListID == listID {
			filtered = append(filtered, item)
			indices = append(indices, i)
		}
	}
	m.filteredItems = filtered
	m.filteredItemIndices = indices
	m.filteredListID = listID
	m.cacheValid = true
	return filtered
}

func (m *model) invalidateCache() {
	m.cacheValid = false
}

func (m *model) getVisibleItemCount() int {
	return len(m.filterItemsByList(m.currentListID))
}

func (m *model) getVisibleItemActualIndex(visibleIndex int) int {
	_ = m.filterItemsByList(m.currentListID) // Ensure cache is populated
	if visibleIndex < 0 || visibleIndex >= len(m.filteredItemIndices) {
		return -1
	}
	return m.filteredItemIndices[visibleIndex]
}

func (m *model) countTasksInList(listID int) int {
	return len(m.filterItemsByList(listID))
}

// renderSyncStatus renders the sync status indicator
func (m *model) renderSyncStatus() string {
	var statusLine string

	if !m.syncEnabled {
		return ""
	}

	// Determine status indicator
	if m.syncStatus.syncing {
		statusLine = "↻ Syncing..."
	} else if !m.syncStatus.online {
		statusLine = "● Offline (local mode)"
	} else if m.syncStatus.errorMessage != "" {
		statusLine = "✗ Sync error: " + m.syncStatus.errorMessage
	} else if m.syncStatus.lastSyncTime > 0 {
		timeSince := time.Since(time.Unix(m.syncStatus.lastSyncTime, 0))
		statusLine = fmt.Sprintf("✓ Synced %s ago", formatDuration(timeSince))
	} else {
		statusLine = "✓ Ready to sync"
	}

	return TitleStyle.Render(statusLine)
}

// formatDuration formats a duration in a user-friendly way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours()) / 24
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
