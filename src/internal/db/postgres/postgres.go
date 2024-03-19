package custpg

import (
	"context"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewSqlx(ctx context.Context, options ...Optioner) (*sqlx.DB, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	globalConfigs := opts.globalConfigs

	logger.SInfo("db.postgres.Init: create SQLx connection",
		zap.String("connection", globalConfigs.Connection))

	client, err := sqlx.Connect("postgres", globalConfigs.Connection)
	if err != nil {
		logger.SFatal("db.postgres.Init: open error",
			zap.Error(err))
		return nil, err
	}

	return client, nil
}

func NewGorm(ctx context.Context, options ...Optioner) (*gorm.DB, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	connString := opts.globalConfigs.Connection
	db, err := gorm.Open(
		postgres.Open(connString),
		&gorm.Config{})
	if err != nil {
		return nil, custerror.FormatInternalError("buildGorm: err = %s", err)
	}

	return db, nil
}

type Options struct {
	globalConfigs *configs.DatabaseConfigs
}

type Optioner func(*Options)

func WithConfigs(globalConfigs *configs.DatabaseConfigs) Optioner {
	return func(o *Options) {
		o.globalConfigs = globalConfigs
	}
}
