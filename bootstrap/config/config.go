package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/bytedance/sonic"
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
	Config  AppConfig
	once    sync.Once
	initErr error
)

const (
	defaultConfigPath = "config/local.json"
)

func InitConfig() error {
	fmt.Fprintf(os.Stdout, "INFO: config: init: started\n")

	if err := InitConfigWithPath(defaultConfigPath); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: config: init: failed, reason=init, error=%v\n", err)
		return fmt.Errorf("config: initialization failed: %w", err)
	}

	fmt.Fprintf(os.Stdout, "INFO: config: init: succeeded, path=%s\n", defaultConfigPath)
	return nil
}

func InitConfigWithPath(configPath string) error {
	once.Do(func() {
		workingDir, err := os.Getwd()
		if err != nil {
			initErr = fmt.Errorf("get working directory failed: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: config: load: failed, reason=get working directory, error=%v\n", err)
			return
		}

		configFilePath := filepath.Join(workingDir, configPath)
		fileBytes, err := os.ReadFile(configFilePath)
		if err != nil {
			initErr = fmt.Errorf("read config file %s failed: %w", configFilePath, err)
			fmt.Fprintf(os.Stderr, "ERROR: config: load: failed, reason=read file, path=%s, error=%v\n", configFilePath, err)
			return
		}

		if err := sonic.Unmarshal(fileBytes, &Config); err != nil {
			initErr = fmt.Errorf("unmarshal config file %s failed: %w", configFilePath, err)
			fmt.Fprintf(os.Stderr, "ERROR: config: load: failed, reason=unmarshal config, path=%s, error=%v\n", configFilePath, err)
			return
		}
	})

	return initErr
}
