package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mergestat/timediff"
)

func (m model) View() string {
	s := []string{TitleStyle.Render("Todo list:")}

	// Update viewport content
	m.updateViewport()
	s = append(s, m.viewport.View())

	if m.editMode {
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
	s = append(s, "Press a to add, e to edit, t to set due date, d to delete, q to quit.\n")
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

	// Adjust scroll offset to keep cursor visible
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	tasksPerView := viewportHeight / LinesPerTask
	if m.cursor >= m.scrollOffset+tasksPerView {
		m.scrollOffset = m.cursor - tasksPerView + 1
	}

	var taskLines []string

	for i, item := range m.items {
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
	if m.editMode || m.inputActive {
		additionalHeight = 6
		if m.inputMode == InputModePriority {
			additionalHeight = 11
		}
	} else if m.deleteConfirm {
		additionalHeight = 2
	}

	return availableHeight - additionalHeight
}
