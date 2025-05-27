package order

import (
	"context"
	"time"

	"github.com/Cora23tt/order_service/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Order struct {
	ID           int
	UserID       int
	Status       string
	DeliveryDate *time.Time
	PickupPoint  string
	OrderDate    time.Time
	TotalAmount  int
	ReceiptURL   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Items        []OrderItem
}

type OrderItem struct {
	ProductID  int
	Quantity   int
	Price      int
	TotalPrice int
}

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, o *Order) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var orderID int
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, status, delivery_date, pickup_point, total_amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, o.UserID, o.Status, o.DeliveryDate, o.PickupPoint, o.TotalAmount).Scan(&orderID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23503": // foreign key violation
				return 0, errors.ErrInvalidInput
			case "23505": // unique violation (если появится)
				return 0, errors.ErrAlreadyExists
			default:
				return 0, errors.ErrInternal
			}
		}
		return 0, err
	}

	for _, item := range o.Items {
		_, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok {
				if pgErr.Code == "23503" {
					return 0, errors.ErrInvalidInput
				}
			}
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (r *Repo) GetByID(ctx context.Context, orderID int) (*Order, error) {
	var o Order
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, delivery_date, pickup_point, order_date, total_amount, receipt_url, created_at, updated_at
		FROM orders WHERE id = $1
	`, orderID).Scan(&o.ID, &o.UserID, &o.Status, &o.DeliveryDate, &o.PickupPoint, &o.OrderDate, &o.TotalAmount, &o.ReceiptURL, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT product_id, quantity, price, total_price
		FROM order_items
		WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price, &item.TotalPrice)
		if err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	return &o, nil
}

func (r *Repo) GetAllByUser(ctx context.Context, userID int) ([]*Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, status, delivery_date, pickup_point, order_date, total_amount, receipt_url, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY order_date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.DeliveryDate, &o.PickupPoint, &o.OrderDate, &o.TotalAmount, &o.ReceiptURL, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, orderID int, status string) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, orderID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, orderID int) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}
