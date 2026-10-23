// Package postgres — реализации репозиториев из usecase/port на pgx.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"smartcredit/internal/domain"
)

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type DB struct{ pool *pgxpool.Pool }

func Connect(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, pool.Ping(ctx)
}

func (db *DB) Close()                         { db.pool.Close() }
func (db *DB) Ping(ctx context.Context) error { return db.pool.Ping(ctx) }

// q — транзакция из ctx, если она есть, иначе пул.
func (db *DB) q(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.pool
}

// notFound переводит pgx.ErrNoRows в доменную ошибку.
func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

// notFoundAs — то же, но с заменой ошибки «нет строки» на as.
func notFoundAs(err, as error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return as
	}
	return err
}
