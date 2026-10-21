package port

import "context"

// TxManager выполняет fn в одной транзакции; репозитории берут её из ctx.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
