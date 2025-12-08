package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Styling constants
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

// Input modes
const (
	InputModeTask     = "task"
	InputModePriority = "priority"
	InputModeDueDate  = "duedate"
	InputModeListName = "listname"
)

// Priority levels
const (
	PriorityHigh    = 1
	PriorityMedHigh = 2
	PriorityMed     = 3
	PriorityLow     = 4
)

// Keyboard shortcuts
const (
	KeyUp    = "up"
	KeyDown  = "down"
	KeyEnter = "enter"
	KeyEsc   = "esc"
	KeyQ     = "q"
	KeyCtrlC = "ctrl+c"
	KeySpace = " "
	KeyK     = "k"
	KeyJ     = "j"
	KeyA     = "a"
	KeyE     = "e"
	KeyD     = "d"
	KeyT     = "t"
	KeyY     = "y"
	KeyN     = "n"
	KeyW     = "w"
	KeyR     = "r"
	KeyL     = "l"
	KeyM     = "m"
)

// Priority selection keys
const (
	KeyPriority1 = "1"
	KeyPriority2 = "2"
	KeyPriority3 = "3"
	KeyPriority4 = "4"
)

// Viewport configuration
const (
	ViewportWidth      = 80
	ViewportHeight     = 20
	ViewportBaseLines  = 5
	LinesPerTask       = 3
	// Viewport height adjustments for different UI modes
	HeightAdjustDefault     = 6    // Standard modal (edit, new task, due date)
	HeightAdjustPriority    = 11   // Priority selection
	HeightAdjustDelete      = 2    // Delete confirmation
	HeightAdjustListPerItem = 1    // Per list item in list selector
)

// Text input configuration
const (
	TextInputCharLimit = 156
	TextInputWidth     = 50
)

// Default values
const (
	DefaultPriority = PriorityMed
)

// Time calculations
const (
	SecondsPerDay = 24 * 3600
)

// Priority style and label maps (initialized in init())
var (
	PriorityStyles map[int]lipgloss.Style
	PriorityLabels map[int]string
	SelectedPriorityStyles map[int]lipgloss.Style
)

func init() {
	PriorityStyles = map[int]lipgloss.Style{
		PriorityHigh:    OneStyle,
		PriorityMedHigh: TwoStyle,
		PriorityMed:     ThreeStyle,
		PriorityLow:     FourStyle,
	}

	SelectedPriorityStyles = map[int]lipgloss.Style{
		PriorityHigh:    SelectedOneStyle,
		PriorityMedHigh: SelectedTwoStyle,
		PriorityMed:     SelectedThreeStyle,
		PriorityLow:     SelectedFourStyle,
	}

	PriorityLabels = map[int]string{
		PriorityHigh:    "🟥 High",
		PriorityMedHigh: "🟧 Medium-High",
		PriorityMed:     "🟨 Medium",
		PriorityLow:     "🟩 Low",
	}
}
