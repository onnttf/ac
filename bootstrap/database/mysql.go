package database

import (
	"fmt"
	"os"
	"sync"
	"time"

	"ac/bootstrap/config"
	"ac/bootstrap/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	DB      *gorm.DB
	once    sync.Once
	initErr error
)

// InitMySQL initializes a global Gorm MySQL connection and verifies connectivity.
func InitMySQL() error {
	once.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: database: init: started\n")

		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			config.Config().Database.User,
			config.Config().Database.Password,
			config.Config().Database.Host,
			config.Config().Database.Port,
			config.Config().Database.Database,
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.NewGormLogger(gormlogger.Info),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to MySQL database: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: database: connect: failed, reason=connect, error=%v\n", err)
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: database: get-sqldb: failed, reason=get db, error=%v\n", err)
			return
		}

		var pingErr error
		for i := 0; i < 3; i++ {
			pingErr = sqlDB.Ping()
			if pingErr == nil {
				break
			}
			fmt.Fprintf(os.Stderr, "WARN: database: ping: retry=%d, error=%v\n", i+1, pingErr)
			time.Sleep([]time.Duration{500 * time.Millisecond, time.Second, 2 * time.Second}[i])
		}
		if pingErr != nil {
			initErr = fmt.Errorf("ping to MySQL failed: %w", pingErr)
			fmt.Fprintf(os.Stderr, "ERROR: database: ping: failed, reason=ping, error=%v\n", pingErr)
			return
		}

		DB = db
		fmt.Fprintf(
			os.Stdout,
			"INFO: database: init: succeeded, host=%s, port=%s, db=%s\n",
			config.Config().Database.Host,
			config.Config().Database.Port,
			config.Config().Database.Database,
		)
	})

	return initErr
}
