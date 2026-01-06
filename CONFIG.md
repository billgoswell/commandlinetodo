# Configuration

This document describes how to configure the Command Line Todo application.

## Database Path

The application stores all todo lists and tasks in a SQLite database. By default, the database is created at `./todo.db` (in the current working directory where the application is run).

### Customizing the Database Path

You can specify a custom database location by setting the `TODO_DB_PATH` environment variable:

**Linux/macOS:**
```bash
export TODO_DB_PATH=/path/to/my/database.db
./commandlinetodo
```

**Windows (PowerShell):**
```powershell
$env:TODO_DB_PATH="C:\path\to\my\database.db"
.\commandlinetodo.exe
```

**Windows (Command Prompt):**
```cmd
set TODO_DB_PATH=C:\path\to\my\database.db
commandlinetodo.exe
```

### Examples

**Using a database in your home directory:**
```bash
export TODO_DB_PATH=~/.todo/tasks.db
./commandlinetodo
```

**Using a database in a shared location:**
```bash
export TODO_DB_PATH=/shared/data/todos.db
./commandlinetodo
```

**Temporary database (cleared on exit):**
```bash
export TODO_DB_PATH=/tmp/todo-session.db
./commandlinetodo
```

## Default Behavior

If the `TODO_DB_PATH` environment variable is not set, the application will:
1. Create the database at `./todo.db` in the current directory
2. Automatically create parent directories if they don't exist
3. Initialize the database schema on first run

## Notes

- The application requires write permissions to the database directory
- The database file is created automatically on first run if it doesn't exist
- The parent directory for the database path is created automatically if needed
- SQLite database files are portable - you can copy and move them as needed
