package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/uptrace/bun"
)

func (d *Driver) RunInTx(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
	return d.RunInTxWithLock(ctx, "", fn)
}

func (d *Driver) RunInTxWithLock(ctx context.Context, lockName string, fn func(context.Context) (any, error)) (any, error) {
	var result any
	err := d.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if lockName != "" {
			key := advisoryLockKey(lockName)
			var ok bool
			if err := tx.QueryRowContext(ctx,
				"SELECT pg_try_advisory_lock(?)", key,
			).Scan(&ok); err != nil {
				return fmt.Errorf("acquire advisory lock: %w", err)
			}
			if !ok {
				return fmt.Errorf("%w: lock %q is already held", store.ErrLocked, lockName)
			}
		}
		res, err := fn(withIDB(ctx, tx))
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
