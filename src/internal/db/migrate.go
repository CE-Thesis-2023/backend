package custdb

import (
	_ "github.com/glebarez/go-sqlite"
	"gorm.io/gorm"
)

func Migrate(gormDb *gorm.DB, schemas ...interface{}) error {
	return gormDb.AutoMigrate(schemas...)
}
