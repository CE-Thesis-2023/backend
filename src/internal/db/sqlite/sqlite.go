package custsqlite

import (
	"context"
	"os"
	"path"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"

	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqlx(ctx context.Context, options ...Optioner) (*sqlx.DB, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	globalConfigs := opts.globalConfigs
	log := logger.Sugar()

	log.Infof("db.sqlite.Init: creating database dsn = %s", globalConfigs.Connection)

	if err := createIfNotExists(globalConfigs.Connection); err != nil {
		log.Fatal(err)
		return nil, err
	}

	client, err := sqlx.Connect("sqlite", globalConfigs.Connection)
	if err != nil {
		log.Fatalf("db.sqlite.Init: open err = %s", err)
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
		sqlite.Open(connString),
		&gorm.Config{})
	if err != nil {
		return nil, custerror.FormatInternalError("buildGorm: err = %s", err)
	}

	return db, nil
}

func createIfNotExists(p string) error {
	fs, err := os.Stat(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return custerror.FormatInternalError("db.sqlite.createIfNotExists: os.Stat err = %s", err)
		}
	}

	if fs != nil {
		return nil
	}
	dir := path.Dir(p)

	os.MkdirAll(dir, 0755)
	if _, err := os.Create(p); err != nil {
		return custerror.FormatInternalError("db.sqlite.createIfNotExists: os.Create err = %s", err)
	}

	return nil
}

type Options struct {
	globalConfigs *configs.DatabaseConfigs
}

type Optioner func(opts *Options)

func WithGlobalConfigs(configs *configs.DatabaseConfigs) Optioner {
	return func(opts *Options) {
		opts.globalConfigs = configs
	}
}
