package main

import (
	"os"
	"path/filepath"
)

type Config struct {
	DBPath string
}

// defaultDBPath is the default location for the database file
const defaultDBPath = "./todo.db"

// dbPathEnvVar is the environment variable name for database path configuration
const dbPathEnvVar = "TODO_DB_PATH"

func LoadConfig() Config {
	cfg := Config{
		DBPath: defaultDBPath,
	}

	if envPath := os.Getenv(dbPathEnvVar); envPath != "" {
		cfg.DBPath = envPath
	}

	if err := ensureDBDirectory(cfg.DBPath); err != nil {
		logErrorMsg("create database directory", err)
	}

	return cfg
}

func ensureDBDirectory(dbPath string) error {
	dir := filepath.Dir(dbPath)

	if dir == "." || dir == "" {
		return nil
	}

	if _, err := os.Stat(dir); err == nil {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}
