package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3c71A8")).
			MarginBottom(1)
	SelectedStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#A3A3A3"))
	DoneStyle         = lipgloss.NewStyle().Strikethrough(true)
	SelectedDoneStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#A3A3A3")).
				Strikethrough(true)
	ItemStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#3c71A8"))
	DocStyle         = lipgloss.NewStyle().Margin(0)
	OneStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	TwoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	ThreeStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	FourStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#008000"))
	SelectedOneStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#A3A3A3")).
				Foreground(lipgloss.Color("#FF0000"))
	SelectedTwoStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#A3A3A3")).
				Foreground(lipgloss.Color("#FFA500"))
	SelectedThreeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#A3A3A3")).
				Foreground(lipgloss.Color("#FFFF00"))
	SelectedFourStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#A3A3A3")).
				Foreground(lipgloss.Color("#008000"))
	OverdueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))
)
