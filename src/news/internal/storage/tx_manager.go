package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/zhavkk/news-service/src/news/internal/config"
)

type TxManagerInterface interface {
	RunSerializable(ctx context.Context, f func(ctx context.Context) error) error
	RunReadUncommited(ctx context.Context, f func(context.Context) error) error
	RunReadCommited(ctx context.Context, f func(context.Context) error) error
	RunRepeatableRead(ctx context.Context, f func(context.Context) error) error
}

type TxManager struct {
	db *Storage
}

func NewTxManager(ctx context.Context, cfg *config.Config) (*TxManager, error) {
	db, err := NewStorage(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("NewTxManager: %w", ErrFailedToConnectToDB)
	}
	return &TxManager{db: db}, nil
}

func (m *TxManager) GetDatabase() *Storage {
	return m.db
}

func (m *TxManager) RunSerializable(ctx context.Context, f func(context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
	return m.beginFunc(ctx, opts, f)
}

func (m *TxManager) RunReadUncommited(ctx context.Context, f func(context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadOnly,
	}
	return m.beginFunc(ctx, opts, f)
}

func (m *TxManager) RunReadCommited(ctx context.Context, f func(ctx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}
	return m.beginFunc(ctx, opts, f)
}

func (m *TxManager) RunRepeatableRead(ctx context.Context, f func(context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	}
	return m.beginFunc(ctx, opts, f)
}
func (m *TxManager) beginFunc(ctx context.Context, opts pgx.TxOptions, f func(context.Context) error) error {
	tx, err := m.db.GetPool().BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctx = context.WithValue(ctx, txKey{}, tx)

	if err := f(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type txKey struct{}

func GetTxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
