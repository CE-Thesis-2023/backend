package custdb

import (
	"context"
	"sync"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custpg "github.com/CE-Thesis-2023/backend/src/internal/db/postgres"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once sync.Once

var layeredDb *LayeredDb
var gormDb *gorm.DB

func Init(ctx context.Context, c *configs.Configs) {
	once.Do(func() {
		layeredDb = NewLayeredDb(ctx, c)
		gormClient, err := custpg.NewGorm(ctx, custpg.WithConfigs(&c.Database))
		if err != nil {
			logger.SFatal("Unable to initialize GORM database", zap.Error(err))
		}
		gormDb = gormClient
	})
}

func Layered() *LayeredDb {
	return layeredDb
}

func Gorm() *gorm.DB {
	return gormDb
}

func Stop(ctx context.Context) {
	layeredDb.sqldb.Close()
}
