package custdb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custpg "github.com/CE-Thesis-2023/backend/src/internal/db/postgres"
	custerror "github.com/CE-Thesis-2023/backend/src/internal/error"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type LayeredDb struct {
	sqldb *sqlx.DB
}

func NewLayeredDb(ctx context.Context, c *configs.Configs) *LayeredDb {
	pg, err := custpg.NewSqlx(ctx, custpg.WithConfigs(&c.Database))
	if err != nil {
		logger.SFatal("Unable to initialize database", zap.Error(err))
	}
	return &LayeredDb{
		sqldb: pg,
	}
}

func (db *LayeredDb) Select(ctx context.Context, selectBuilder sq.SelectBuilder, dest interface{}) error {
	queryStr, arguments, err := selectBuilder.ToSql()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Select: ToSql() err = %s", err)
	}

	logger.SDebug("LayeredDb.Select",
		zap.String("query", queryStr),
		zap.Reflect("arguments", arguments))

	sqldb := db.sqldb

	err = sqldb.SelectContext(ctx, dest, queryStr, arguments...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return custerror.ErrorNotFound
		}
		return custerror.FormatInternalError("LayeredDb.Select: SelectContext err = %s", err)
	}
	return nil
}

func (db *LayeredDb) Get(ctx context.Context, selectBuilder sq.SelectBuilder, dest interface{}) error {
	queryStr, arguments, err := selectBuilder.ToSql()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Select: ToSql() err = %s", err)
	}

	logger.SDebug("LayeredDb.Get",
		zap.String("query", queryStr),
		zap.Reflect("arguments", arguments))

	sqldb := db.sqldb

	err = sqldb.GetContext(ctx, dest, queryStr, arguments...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return custerror.ErrorNotFound
		}
		return custerror.FormatInternalError("LayeredDb.Get: GetContext err = %s", err)
	}

	return nil
}

func (db *LayeredDb) Insert(ctx context.Context, insertBuilder sq.InsertBuilder) error {
	queryStr, arguments, err := insertBuilder.ToSql()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Insert: ToSql() err = %s", err)
	}

	logger.SDebug("LayeredDb.Insert",
		zap.String("query", queryStr),
		zap.Reflect("arguments", arguments))

	sqldb := db.sqldb

	_, err = sqldb.ExecContext(ctx, queryStr, arguments...)
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Insert: Exec() err = %s", err)
	}

	return nil
}

func (db *LayeredDb) Update(ctx context.Context, updateBuilder sq.UpdateBuilder) error {
	queryStr, arguments, err := updateBuilder.ToSql()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Update: ToSql() err = %s", err)
	}

	logger.SDebug("LayeredDb.Update",
		zap.String("query", queryStr),
		zap.Reflect("arguments", arguments))

	sqldb := db.sqldb

	res, err := sqldb.ExecContext(ctx, queryStr, arguments...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return custerror.ErrorNotFound
		}
		return custerror.FormatInternalError("LayeredDb.Update: Exec() err = %s", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Update: RowsAffected() err = %s", err)
	}

	if affected == 0 {
		return custerror.ErrorNotFound
	}

	return nil
}

func (db *LayeredDb) Delete(ctx context.Context, deleteBuilder sq.DeleteBuilder) error {
	queryStr, arguments, err := deleteBuilder.ToSql()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Delete: ToSql() err = %s", err)
	}

	logger.SDebug("LayeredDb.Delete",
		zap.String("query", queryStr),
		zap.Reflect("arguments", arguments))

	sqldb := db.sqldb

	res, err := sqldb.ExecContext(ctx, queryStr, arguments...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return custerror.ErrorNotFound
		}
		return custerror.FormatInternalError("LayeredDb.Delete: Exec() err = %s", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return custerror.FormatInternalError("LayeredDb.Delete: RowsAffected() err = %s", err)
	}

	if affected == 0 {
		return custerror.ErrorNotFound
	}

	return nil
}
