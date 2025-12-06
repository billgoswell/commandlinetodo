# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a command-line todo list application written in Go. It provides a terminal UI for managing tasks with priorities and completion status. The application uses SQLite for persistence and Bubble Tea framework for interactive terminal UI.

## Architecture

### Directory Structure

```
cmd/app/
├── main.go           # Application entry point
├── model.go          # Data model and state management
├── handlers.go       # Keyboard/input event handling
├── render.go         # UI rendering and display
├── db.go             # Database operations and schemas
├── utils.go          # Utility functions (date parsing)
└── constants.go      # Styling, keyboard shortcuts, and constants
```

### Core Components

1. **TUI Framework**: Built with Charmbracelet libraries:
   - `bubbletea` - Main terminal UI framework
   - `lipgloss` - Terminal styling and layout
   - `bubbles` - Reusable components (textinput, viewport)

2. **Data Layer**:
   - SQLite database (`todo.db`) with two tables:
     - `todolists` - Manages multiple todo lists/workspaces
     - `tasks` - Individual tasks linked to todolists
   - Full CRUD operations for both tables

3. **Model Structure** (`model.go`):
   - `model` struct contains application state
   - Tracks current todolist and cursor position
   - Uses Bubble Tea's Update/View pattern

4. **Styling** (`constants.go`):
   - Style definitions for priorities (colors 1-4)
   - Keyboard shortcuts and constants
   - Input modes and view states

### Key Data Structures

**todoitem** struct (db.go):
- `id`, `done`, `todo`, `priority`, `duedate`
- `dateadded`, `datecompleted` - Unix timestamps
- `todolistID` - Links to parent todolist
- `deleted`, `deletedat` - Soft delete support

**todolist** struct (db.go):
- `id`, `name`, `displayOrder`
- `archived` - Hide without deleting
- `createdAt`, `updatedAt` - Unix timestamps

## Build & Run

```bash
# Build the application
go build -o commandlinetodo ./cmd/app

# Run the application
./commandlinetodo

# Run tests
go test ./...
```

## Current Features

### Task Management
- `a` - Add new task
- `e` - Edit task
- `d` - Delete task
- `t` - Set due date
- `Space`/`Enter` - Toggle completion
- `k`/`↑`, `j`/`↓` - Navigate tasks
- `q`/`Ctrl+C` - Quit

### Todo List Management (Planned)
- `l` - Open list selector modal
- `n` - Create new list
- List rename, delete, and archive (in development)

## Development Notes

### Completed Features
- ✅ Full SQLite database integration
- ✅ Task CRUD operations with persistence
- ✅ Priority levels (1-4) with color coding
- ✅ Due date support with overdue detection
- ✅ Task soft deletion
- ✅ Code organization in cmd/app structure

### In Development
- 🔄 **Multiple Todo Lists** (Current feature being implemented)
  - Phase 1: Model & database layer
  - Phase 2: List selector UI
  - Phase 3: List creation
  - Phase 4: List management (rename/delete/archive)
  - Phase 5: Data migration for existing tasks

### Known Gaps
- Tasks may have inconsistent `todolist_id` values from old data
- No archive functionality for lists yet
- No list reordering capability

## Implementation Plan: Multiple Todo Lists

See `/home/bill/.claude/plans/buzzing-juggling-volcano.md` for detailed implementation plan.

### Quick Summary
Users can manage multiple todo lists with:
- **Switch lists**: Press `w` to open selector
- **Create list**: Press `n` for new list
- **Manage**: Rename, delete, archive lists
- **Persistence**: All changes saved to SQLite

### Files Being Modified
- `cmd/app/model.go` - Add todolist tracking
- `cmd/app/db.go` - Add list CRUD functions
- `cmd/app/handlers.go` - Keyboard handlers for lists
- `cmd/app/render.go` - List selector UI
- `cmd/app/constants.go` - New keyboard shortcuts
- `cmd/app/main.go` - Load todolists on startup
