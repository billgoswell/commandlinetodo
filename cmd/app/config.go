package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DBPath string
	Sync   SyncConfig
}

type SyncConfig struct {
	Enabled             bool
	ServerURL           string
	APIKey              string
	DeviceID            string
	SyncIntervalSeconds int
	AutoSyncOnChange    bool
	RetryAttempts       int
	TimeoutSeconds      int
}

// defaultDBPath is the default location for the database file
const defaultDBPath = "./todo.db"

// dbPathEnvVar is the environment variable name for database path configuration
const dbPathEnvVar = "TODO_DB_PATH"

// Sync environment variables
const (
	syncEnabledEnvVar      = "TODO_SYNC_ENABLED"
	syncServerURLEnvVar    = "TODO_SYNC_SERVER_URL"
	syncAPIKeyEnvVar       = "TODO_SYNC_API_KEY"
	syncDeviceIDEnvVar     = "TODO_SYNC_DEVICE_ID"
	syncIntervalEnvVar     = "TODO_SYNC_INTERVAL"
	autoSyncOnChangeEnvVar = "TODO_AUTO_SYNC_ON_CHANGE"
	retryAttemptsEnvVar    = "TODO_SYNC_RETRY_ATTEMPTS"
	timeoutSecondsEnvVar   = "TODO_SYNC_TIMEOUT"
)

// Default sync configuration values
const (
	defaultSyncInterval    = 60
	defaultRetryAttempts   = 3
	defaultTimeoutSeconds  = 10
)

func LoadConfig() Config {
	cfg := Config{
		DBPath: defaultDBPath,
		Sync:   loadSyncConfig(),
	}

	if envPath := os.Getenv(dbPathEnvVar); envPath != "" {
		cfg.DBPath = envPath
	}

	if err := ensureDBDirectory(cfg.DBPath); err != nil {
		logErrorMsg("create database directory", err)
	}

	return cfg
}

func loadSyncConfig() SyncConfig {
	syncCfg := SyncConfig{
		Enabled:             parseBoolEnv(syncEnabledEnvVar, false),
		ServerURL:           os.Getenv(syncServerURLEnvVar),
		APIKey:              os.Getenv(syncAPIKeyEnvVar),
		DeviceID:            os.Getenv(syncDeviceIDEnvVar),
		SyncIntervalSeconds: parseIntEnv(syncIntervalEnvVar, defaultSyncInterval),
		AutoSyncOnChange:    parseBoolEnv(autoSyncOnChangeEnvVar, true),
		RetryAttempts:       parseIntEnv(retryAttemptsEnvVar, defaultRetryAttempts),
		TimeoutSeconds:      parseIntEnv(timeoutSecondsEnvVar, defaultTimeoutSeconds),
	}

	// Generate device ID if not set
	if syncCfg.DeviceID == "" {
		syncCfg.DeviceID = generateClientID()
	}

	return syncCfg
}

func parseBoolEnv(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	// Accept "true", "1", "yes" as true
	return strings.ToLower(val) == "true" || val == "1" || strings.ToLower(val) == "yes"
}

func parseIntEnv(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}
	return defaultVal
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
