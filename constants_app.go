package main

// Input modes
const (
	InputModeTask     = "task"
	InputModePriority = "priority"
	InputModeDueDate  = "duedate"
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
	ViewportWidth     = 80
	ViewportHeight    = 20
	ViewportBaseLines = 5
	LinesPerTask      = 3
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
