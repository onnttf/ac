package database

import (
	"fmt"
	"os"
	"sync"

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

func InitMySQL() error {
	once.Do(func() {
		fmt.Fprintf(os.Stdout, "INFO: database: init: started\n")

		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
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

		if err := sqlDB.Ping(); err != nil {
			initErr = fmt.Errorf("ping to MySQL failed: %w", err)
			fmt.Fprintf(os.Stderr, "ERROR: database: ping: failed, reason=ping, error=%v\n", err)
			return
		}

		DB = db
		fmt.Fprintf(
			os.Stdout,
			"INFO: database: init: succeeded, host=%s, port=%s, db=%s\n",
			config.Config.Database.Host,
			config.Config.Database.Port,
			config.Config.Database.Database,
		)

		// g := gen.NewGenerator(gen.Config{
		// 	ModelPkgPath:      "./model",
		// 	FieldWithIndexTag: true,
		// 	FieldWithTypeTag:  true,
		// })
		// var dataMap = map[string]func(gorm.ColumnType) (dataType string){
		// 	"int": func(columnType gorm.ColumnType) (dataType string) {
		// 		return "int64"
		// 	},
		// 	"tinyint": func(columnType gorm.ColumnType) (dataType string) {
		// 		return "int64"
		// 	},
		// }
		// g.WithDataTypeMap(dataMap)
		// g.UseDB(DB)
		// g.GenerateAllTable()
		// g.Execute()
	})

	return initErr
}
