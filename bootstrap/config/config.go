package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
}

type LogConfig struct {
	Level     string `json:"level"`
	Directory string `json:"directory"`
}

type AppConfig struct {
	Database DatabaseConfig `json:"database"`
	Log      LogConfig      `json:"log"`
}

var (
	Config AppConfig
	once   sync.Once
)

// InitConfig initializes the Config from the config file
func InitConfig() error {
	var initErr error
	once.Do(func() {
		workingDir, err := os.Getwd()
		if err != nil {
			initErr = fmt.Errorf("failed to get current working directory, err: %w", err)
			return
		}

		configFilePath := filepath.Join(workingDir, "config", "local.json")

		fileBytes, err := os.ReadFile(configFilePath)
		if err != nil {
			initErr = fmt.Errorf("failed to read config file, err: %w", err)
			return
		}

		if err := json.Unmarshal(fileBytes, &Config); err != nil {
			initErr = fmt.Errorf("failed to unmarshal config file, err: %w", err)
			return
		}
	})
	return initErr
}
