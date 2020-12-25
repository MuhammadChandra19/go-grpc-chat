package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type contextKey int

// List of context keys for user context.
const (
	contextKeyTx contextKey = iota
)

// NewContextTx creates a new context with the *sqlx.Tx value.
func NewContextTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	ctx = context.WithValue(ctx, contextKeyTx, tx)
	return ctx
}

// TxFromContext gets the *sqlx.Tx value from the context.
func TxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(contextKeyTx).(*sqlx.Tx)
	return tx, ok
}
