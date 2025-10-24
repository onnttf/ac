package bootstrap

import (
	"fmt"
	"os"

	"ac/bootstrap/config"
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
)

func Initialize() error {
	fmt.Fprintf(os.Stdout, "INFO: bootstrap: application initializing\n")

	if err := config.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: bootstrap: config initialization failed, error=%v\n", err)
		return fmt.Errorf("bootstrap: failed to initialize config: %w", err)
	}

	logConfig := logger.Config{
		Directory: config.Config.Log.Directory,
		Level:     config.Config.Log.Level,
	}
	if err := logger.InitLogger(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: bootstrap: logger initialization failed, error=%v\n", err)
		return fmt.Errorf("bootstrap: failed to initialize logger: %w", err)
	}

	if err := database.InitMySQL(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: bootstrap: database initialization failed, error=%v\n", err)
		return fmt.Errorf("bootstrap: failed to initialize database: %w", err)
	}

	fmt.Fprintf(os.Stdout, "INFO: bootstrap: application initialized successfully\n")
	return nil
}
