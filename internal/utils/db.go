package utils

import (
	"sync"

	"github.com/cicbyte/memos-cli/internal/log"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	gormDB *gorm.DB
	dbOnce sync.Once
)

func GetGormDB() (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		dbPath := ConfigInstance.GetDbPath()
		gormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: log.GetGormLogger(),
		})
	})
	if err != nil {
		return nil, err
	}
	return gormDB, err
}

func CloseGormDB() error {
	if gormDB != nil {
		sqlDB, err := gormDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
