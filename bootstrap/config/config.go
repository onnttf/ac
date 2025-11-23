package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bytedance/sonic"
)

type (
	DatabaseConfig struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
	}

	LogConfig struct {
		Level     string `json:"level"`
		Directory string `json:"directory"`
	}

	AppConfig struct {
		Database DatabaseConfig `json:"database"`
		Log      LogConfig      `json:"log"`
	}
)

var cfg AppConfig

const (
	defaultConfigPath  = "config/local.json"
	fallbackConfigPath = "config/example.json"
)

func Config() AppConfig { return cfg }

func InitConfig() error {
	fmt.Fprintf(os.Stdout, "INFO: config: init: started\n")

	if envPath := os.Getenv("APP_CONFIG"); envPath != "" {
		if err := InitConfigWithPath(envPath); err == nil {
			fmt.Fprintf(os.Stdout, "INFO: config: init: succeeded, path=%s\n", envPath)
			return nil
		}
		fmt.Fprintf(os.Stderr, "WARN: config: init: env APP_CONFIG failed, path=%s\n", envPath)
	}

	if err := InitConfigWithPath(defaultConfigPath); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: config: init: default path failed, path=%s, error=%v\n", defaultConfigPath, err)
		if err2 := InitConfigWithPath(fallbackConfigPath); err2 != nil {
			fmt.Fprintf(os.Stderr, "ERROR: config: init: failed, error=%v\n", err2)
			return fmt.Errorf("config: initialization failed: %w", err2)
		}
		fmt.Fprintf(os.Stdout, "INFO: config: init: succeeded, path=%s\n", fallbackConfigPath)
		return nil
	}

	fmt.Fprintf(os.Stdout, "INFO: config: init: succeeded, path=%s\n", defaultConfigPath)
	return nil
}

func InitConfigWithPath(configPath string) error {
	var fileBytes []byte

	if filepath.IsAbs(configPath) {
		b, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("read config file %s failed: %w", configPath, err)
		}
		fileBytes = b
	} else {
		workingDir, err := os.Getwd()
		if err == nil {
			p := filepath.Join(workingDir, configPath)
			if b, err := os.ReadFile(p); err == nil {
				fileBytes = b
			}
		}
		if len(fileBytes) == 0 {
			exe, _ := os.Executable()
			base := filepath.Dir(exe)
			p := filepath.Join(base, configPath)
			b, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read config file %s failed: %w", p, err)
			}
			fileBytes = b
		}
	}

	var newCfg AppConfig
	if err := sonic.Unmarshal(fileBytes, &newCfg); err != nil {
		return fmt.Errorf("unmarshal config failed: %w", err)
	}

	if newCfg.Log.Level == "" {
		newCfg.Log.Level = "info"
	}
	if newCfg.Log.Directory == "" {
		newCfg.Log.Directory = "log"
	}

	if err := newCfg.Validate(); err != nil {
		return err
	}

	cfg = newCfg
	return nil
}

func (c AppConfig) Validate() error {
	if c.Database.User == "" || c.Database.Host == "" || c.Database.Port == "" || c.Database.Database == "" {
		return fmt.Errorf("invalid database config")
	}
	if _, err := strconv.Atoi(c.Database.Port); err != nil {
		return fmt.Errorf("invalid database port")
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level")
	}
	if c.Log.Directory == "" {
		return fmt.Errorf("invalid log directory")
	}
	return nil
}
