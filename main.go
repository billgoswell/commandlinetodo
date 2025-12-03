package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
)

type model struct {
	items  []todoitem
	cursor int
	width  int
	height int
	inputs []textinput.Model
}

func main() {
	todoitems, err := getIntialItems()
	if err != nil {
		fmt.Println(err)
		return
	}
	p := tea.NewProgram(intialModel(todoitems), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting tea: %v", err)
	}
	os.Exit(1)
}

func intialModel(todoitems []todoitem) model {
	return model{
		items: todoitems,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items) {
				m.cursor++
			}
		case "enter", " ":
			if m.items[m.cursor].done {
				m.items[m.cursor].done = false
			} else {
				m.items[m.cursor].done = true
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := []string{TitleStyle.Render("Todo list: \n\n")}

	for i, item := range m.items {

		c := fmt.Sprintf("%s\n%s\n", item.todo, item.notes)
		if m.cursor == i {
			c = SelectedStyle.Render(c)
		}
		if item.done {
			c = DoneStyle.Render(c)
		} else {
			switch item.priority {
			case 1:
				c = OneStyle.Render(c)
			case 2:
				c = TwoStyle.Render(c)
			case 3:
				c = ThreeStyle.Render(c)
			case 4:
				c = FourStyle.Render(c)
			}
		}
		s = append(s, c)

	}
	if m.cursor == len(m.items) {
		s = append(s, SelectedStyle.Render("Enter a new task"))
	} else {
		s = append(s, TitleStyle.Render("Enter a new task"))
	}

	s = append(s, "\nPress q to quit.\n")
	return DocStyle.Render(lipgloss.JoinVertical(lipgloss.Left, s...))
}
