package order

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Cora23tt/order_service/pkg/enums"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repo struct {
	db  TxExecutor
	log *zap.SugaredLogger
}

func NewRepo(db *pgxpool.Pool, log *zap.SugaredLogger) *Repo {
	return &Repo{db: db, log: log}
}

func NewWithTx(tx pgx.Tx, log *zap.SugaredLogger) *Repo {
	return &Repo{db: tx, log: log}
}

type TxExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Order struct {
	ID           int64             `json:"id"`
	UserID       int64             `json:"user_id"`
	Status       enums.OrderStatus `json:"status"`
	DeliveryDate *time.Time        `json:"delivery_date,omitempty"`
	PickupPoint  string            `json:"pickup_point"`
	OrderDate    time.Time         `json:"order_date"`
	TotalAmount  int64             `json:"total_amount"`
	ReceiptURL   *string           `json:"receipt_url,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Items        []OrderItem       `json:"items,omitempty"`
}

type OrderItem struct {
	ProductID  int64 `json:"product_id"`
	Quantity   int64 `json:"quantity"`
	Price      int64 `json:"price"`
	TotalPrice int64 `json:"total_price"`
}

func (r *Repo) Create(ctx context.Context, o *Order) (int64, error) {
	var orderID int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO orders (user_id, status, delivery_date, pickup_point, total_amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, o.UserID, o.Status, o.DeliveryDate, o.PickupPoint, o.TotalAmount).Scan(&orderID)
	if err != nil {
		r.log.Errorw("insert order failed", "userID", o.UserID, "error", err)
		return 0, r.handlePgError(err, "create order")
	}

	for _, item := range o.Items {
		_, err := r.db.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			r.log.Errorw("insert order item failed", "orderID", orderID, "productID", item.ProductID, "error", err)
			return 0, r.handlePgError(err, "insert order item")
		}
	}

	r.log.Infow("order created", "orderID", orderID, "userID", o.UserID)
	return orderID, nil
}

func (r *Repo) GetByID(ctx context.Context, orderID int64) (*Order, error) {
	var o Order
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, delivery_date, pickup_point, order_date, total_amount, receipt_url, created_at, updated_at
		FROM orders WHERE id = $1
	`, orderID).Scan(&o.ID, &o.UserID, &o.Status, &o.DeliveryDate, &o.PickupPoint, &o.OrderDate, &o.TotalAmount, &o.ReceiptURL, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warnw("order not found", "orderID", orderID)
			return nil, pkgerrors.ErrNotFound
		}
		r.log.Errorw("get order failed", "orderID", orderID, "error", err)
		return nil, pkgerrors.ErrInternal
	}

	rows, err := r.db.Query(ctx, `
		SELECT product_id, quantity, price, total_price
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		r.log.Errorw("get order items failed", "orderID", orderID, "error", err)
		return nil, pkgerrors.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price, &item.TotalPrice); err != nil {
			r.log.Errorw("scan order item failed", "orderID", orderID, "error", err)
			return nil, pkgerrors.ErrInternal
		}
		o.Items = append(o.Items, item)
	}

	return &o, nil
}

func (r *Repo) GetAllByUser(ctx context.Context, userID int64) ([]*Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, status, delivery_date, pickup_point, order_date, total_amount, receipt_url, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY order_date DESC
	`, userID)
	if err != nil {
		r.log.Errorw("get all orders failed", "userID", userID, "error", err)
		return nil, pkgerrors.ErrInternal
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.DeliveryDate, &o.PickupPoint, &o.OrderDate, &o.TotalAmount, &o.ReceiptURL, &o.CreatedAt, &o.UpdatedAt); err != nil {
			r.log.Errorw("scan order failed", "userID", userID, "error", err)
			return nil, pkgerrors.ErrInternal
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, orderID)
	if err != nil {
		r.log.Errorw("update order status failed", "orderID", orderID, "status", status, "error", err)
		return r.handlePgError(err, "update status")
	}
	if cmd.RowsAffected() == 0 {
		r.log.Warnw("order not found for update", "orderID", orderID)
		return pkgerrors.ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, orderID int64) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
	if err != nil {
		r.log.Errorw("delete order failed", "orderID", orderID, "error", err)
		return r.handlePgError(err, "delete order")
	}
	if cmd.RowsAffected() == 0 {
		r.log.Warnw("order not found for delete", "orderID", orderID)
		return pkgerrors.ErrNotFound
	}
	r.log.Infow("order deleted", "orderID", orderID)
	return nil
}

func (r *Repo) handlePgError(err error, context string) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pkgerrors.PGErrForeignKeyViolation:
			return pkgerrors.ErrInvalidInput
		case pkgerrors.PGErrUniqueViolation:
			return pkgerrors.ErrAlreadyExists
		case pkgerrors.PGErrInvalidTextRep, pkgerrors.PGErrInvalidType:
			return pkgerrors.ErrInvalidInput
		default:
			r.log.Errorw(context+" failed", "pg_code", pgErr.Code, "pg_msg", pgErr.Message)
			return pkgerrors.ErrInternal
		}
	}
	r.log.Errorw(context+" failed (non-pg)", "error", err)
	return pkgerrors.ErrInternal
}

type OrderStats struct {
	Status enums.OrderStatus `json:"status"`
	Count  int64             `json:"count"`
}

func (r *Repo) GetStats(ctx context.Context, from, to time.Time) ([]OrderStats, error) {
	query := `
		SELECT status, COUNT(*) 
		FROM orders 
		WHERE order_date BETWEEN $1 AND $2
		GROUP BY status
	`
	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		r.log.Errorw("get order stats failed", "error", err)
		return nil, pkgerrors.ErrInternal
	}
	defer rows.Close()

	var stats []OrderStats
	for rows.Next() {
		var s OrderStats
		if err := rows.Scan(&s.Status, &s.Count); err != nil {
			r.log.Errorw("scan order stats failed", "error", err)
			return nil, pkgerrors.ErrInternal
		}
		stats = append(stats, s)
	}
	return stats, nil
}

type ExportFilter struct {
	UserID    *int64
	Status    *enums.OrderStatus
	MinAmount *int64
	MaxAmount *int64
	Limit     int
	Offset    int
}

func (r *Repo) Export(ctx context.Context, f ExportFilter) ([]*Order, error) {
	var (
		query  = `SELECT id, user_id, status, delivery_date, pickup_point, order_date, total_amount, receipt_url, created_at, updated_at FROM orders WHERE 1=1`
		params []interface{}
		index  = 1
	)

	if f.UserID != nil {
		query += ` AND user_id = $` + strconv.Itoa(index)
		params = append(params, *f.UserID)
		index++
	}
	if f.Status != nil {
		query += ` AND status = $` + strconv.Itoa(index)
		params = append(params, *f.Status)
		index++
	}
	if f.MinAmount != nil {
		query += ` AND total_amount >= $` + strconv.Itoa(index)
		params = append(params, *f.MinAmount)
		index++
	}
	if f.MaxAmount != nil {
		query += ` AND total_amount <= $` + strconv.Itoa(index)
		params = append(params, *f.MaxAmount)
		index++
	}

	query += ` ORDER BY order_date DESC LIMIT $` + strconv.Itoa(index)
	params = append(params, f.Limit)
	index++

	query += ` OFFSET $` + strconv.Itoa(index)
	params = append(params, f.Offset)

	rows, err := r.db.Query(ctx, query, params...)
	if err != nil {
		r.log.Errorw("export orders failed", "error", err)
		return nil, pkgerrors.ErrInternal
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.DeliveryDate, &o.PickupPoint, &o.OrderDate, &o.TotalAmount, &o.ReceiptURL, &o.CreatedAt, &o.UpdatedAt); err != nil {
			r.log.Errorw("scan order failed", "error", err)
			return nil, pkgerrors.ErrInternal
		}
		orders = append(orders, &o)
	}

	return orders, nil
}
