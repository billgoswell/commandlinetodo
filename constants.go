package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3c71A8")).
			MarginBottom(1)
	SelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#A3A3A3"))
	DoneStyle     = lipgloss.NewStyle().Strikethrough(true)
	ItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c71A8"))
	DocStyle      = lipgloss.NewStyle().Margin(0)
	OneStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3B41"))
	TwoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC547"))
	ThreeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFf047"))
	FourStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#C2FF47"))
)
