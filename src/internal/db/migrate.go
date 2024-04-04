package custdb

import (
	"gorm.io/gorm"
)

func Migrate(gormDb *gorm.DB, schemas ...interface{}) error {
	return gormDb.AutoMigrate(schemas...)
}
