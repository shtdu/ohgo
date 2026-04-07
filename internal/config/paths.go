package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultBaseDir    = ".openharness"
	configFileName    = "settings.json"
	projectConfigDir  = ".openharness"
	envConfigDir      = "OPENHARNESS_CONFIG_DIR"
	envDataDir        = "OPENHARNESS_DATA_DIR"
	envLogsDir        = "OPENHARNESS_LOGS_DIR"
)

// ConfigDir returns the configuration directory, creating it if needed.
// Resolution order: OPENHARNESS_CONFIG_DIR env var, then ~/.openharness/.
func ConfigDir() (string, error) {
	if dir := os.Getenv(envConfigDir); dir != "" {
		return ensureDir(dir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return ensureDir(filepath.Join(home, defaultBaseDir))
}

// ConfigFilePath returns the path to the main settings file.
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// DataDir returns the data directory for caches, history, etc.
func DataDir() (string, error) {
	if dir := os.Getenv(envDataDir); dir != "" {
		return ensureDir(dir)
	}
	cfgDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(cfgDir, "data"))
}

// LogsDir returns the logs directory.
func LogsDir() (string, error) {
	if dir := os.Getenv(envLogsDir); dir != "" {
		return ensureDir(dir)
	}
	cfgDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(cfgDir, "logs"))
}

// SessionsDir returns the session storage directory.
func SessionsDir() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(dataDir, "sessions"))
}

// TasksDir returns the background task output directory.
func TasksDir() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(dataDir, "tasks"))
}

// FeedbackDir returns the feedback storage directory.
func FeedbackDir() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return ensureDir(filepath.Join(dataDir, "feedback"))
}

// CronRegistryPath returns the cron registry file path.
func CronRegistryPath() (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "cron_jobs.json"), nil
}

// ProjectDir returns the per-project .openharness directory.
func ProjectDir(cwd string) (string, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("cannot resolve cwd: %w", err)
	}
	return ensureDir(filepath.Join(abs, projectConfigDir))
}

func ensureDir(path string) (string, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("cannot create directory %s: %w", path, err)
	}
	return path, nil
}
