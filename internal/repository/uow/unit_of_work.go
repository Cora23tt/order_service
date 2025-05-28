package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork interface {
	Begin(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetTx() pgx.Tx
}

type PgxUnitOfWork struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *PgxUnitOfWork {
	return &PgxUnitOfWork{db: db}
}

func (u *PgxUnitOfWork) Begin(ctx context.Context) (Tx, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTx{tx: tx}, nil
}

type pgxTx struct {
	tx pgx.Tx
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *pgxTx) GetTx() pgx.Tx {
	return t.tx
}
