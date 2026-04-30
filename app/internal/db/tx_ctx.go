package db

import (
	"context"

	"github.com/uptrace/bun"
)

type idbCtxKey struct{}

func withIDB(ctx context.Context, db bun.IDB) context.Context {
	return context.WithValue(ctx, idbCtxKey{}, db)
}

func idbFrom(ctx context.Context, fallback bun.IDB) bun.IDB {
	if v := ctx.Value(idbCtxKey{}); v != nil {
		if db, ok := v.(bun.IDB); ok {
			return db
		}
	}
	return fallback
}
