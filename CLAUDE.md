# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a command-line todo list application written in Go. It provides a terminal UI for managing tasks with priorities and completion status. The application reads tasks from a CSV file (`todo.csv`) and displays them in an interactive interface using the Bubble Tea framework.

## Architecture

### Core Components

1. **TUI Framework**: Built with Charmbracelet libraries:
   - `bubbletea` - Main terminal UI framework
   - `lipgloss` - Terminal styling and layout
   - `bubbles` - Reusable components (e.g., text input)

2. **Data Layer**:
   - Currently reads from `todo.csv` file (via `fileparse.go`)
   - Partial SQLite integration started (`db.go`) but not yet functional
   - CSV format: `id,done,todo_text,priority,notes`

3. **Model Structure** (`main.go`):
   - `model` struct contains the application state (items, cursor position, screen dimensions)
   - Uses Bubble Tea's Update/View pattern for state management and rendering

4. **Styling** (`constants.go`):
   - Style definitions for different task priorities (colors 1-4)
   - Styles for selected items, completed items, and title

### Key Data Structure

`todoitem` struct (db.go):
- `id` (int)
- `done` (bool)
- `todo` (string)
- `priority` (int) - Values 1-4, each with distinct colors
- `notes` (string)
- `datecompleted` (int) - Unix timestamp
- `dateadded` (int) - Unix timestamp

## Build & Run

```bash
# Build the application
go build

# Run the application
./commandlinetodo

# Run tests (none currently exist)
go test ./...
```

## Keyboard Controls

- `q` or `Ctrl+C` - Quit
- `k` or `Up` - Move cursor up
- `j` or `Down` - Move cursor down
- `Space` or `Enter` - Toggle task completion status
- Bottom option "Enter a new task" is selectable but not yet functional

## Development Notes

- **CSV Parsing** (`fileparse.go`): The `getIntialItems()` function reads from `todo.csv` on startup. Error handling currently returns early without persistence.
- **Database** (`db.go`): SQLite table is defined but the `addToDB()` function is not implemented. The database is opened but not currently used.
- **Persistence**: Changes made in the UI are not saved back to the CSV or database - this is a gap in the implementation.
- **New Task Creation**: The UI shows "Enter a new task" as an option but doesn't have input handling implemented for creating new tasks.

## Common Issues & Known Gaps

- No persistence of task state changes
- Database layer incomplete (schema defined but operations not implemented)
- New task creation UI is visible but non-functional
- Text input components imported but not used in the current implementation
