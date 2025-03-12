package database

import (
	"ac/bootstrap/config"
	"ac/bootstrap/logger"
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	DB   *gorm.DB
	once sync.Once
)

// InitMySQL initializes the MySQL database connection using GORM
func InitMySQL() error {
	var initErr error
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			config.Config.Database.User,
			config.Config.Database.Password,
			config.Config.Database.Host,
			config.Config.Database.Port,
			config.Config.Database.Database,
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.NewGormLogger(gormlogger.Info),
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to MySQL, err: %w", err)
			return
		}
		DB = db

		//g := gen.NewGenerator(gen.Config{
		//	ModelPkgPath:      "./model",
		//	FieldNullable:     true,
		//	FieldCoverable:    true,
		//	FieldSignable:     true,
		//	FieldWithIndexTag: true,
		//	FieldWithTypeTag:  true,
		//})
		//
		//g.UseDB(DB) // reuse your gorm db
		//
		//g.GenerateAllTable()
		//// Generate the code
		//g.Execute()
	})
	return initErr
}
