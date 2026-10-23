package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txKey struct{}

// WithinTx реализует port.TxManager. Вложенный вызов переиспользует внешнюю транзакцию.
func (db *DB) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}
