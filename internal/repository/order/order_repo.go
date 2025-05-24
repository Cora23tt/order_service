package order

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	statusPendingPayment = "pending_payment"
	statusPaid           = "paid"
	statusProcessing     = "processing"
	statusShipped        = "shipped"
	statusDelivered      = "delivered"
	statusCancelled      = "cancelled"
)

type Order struct {
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, e *Order) error {
	return nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*Order, error) {
	return &Order{}, nil
}

func (r *Repo) GetAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	return orders, nil
}

func (r *Repo) Update(ctx context.Context, e *Order) error {
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	return nil
}
