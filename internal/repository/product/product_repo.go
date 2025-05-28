package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Product struct {
	ID            int64
	Name          string
	Description   string
	ImageUrl      string
	Price         int64
	StockQuantity int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repo struct {
	db  TxExecutor
	tx  pgx.Tx
	log *zap.SugaredLogger
}

func NewRepo(db *pgxpool.Pool, log *zap.SugaredLogger) *Repo {
	return &Repo{db: db, log: log}
}

func NewWithTx(tx pgx.Tx, log *zap.SugaredLogger) *Repo {
	return &Repo{tx: tx, log: log}
}

type TxExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *Repo) CreateProduct(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (name, description, image_url, price, stock_quantity)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.tx.Exec(ctx, query,
		product.Name,
		product.Description,
		product.ImageUrl,
		product.Price,
		product.StockQuantity,
	)
	r.log.Infow("product created", "name", product.Name)
	return r.handlePgError("create product", err)
}

func (r *Repo) DeleteProduct(ctx context.Context, productID int64) error {
	cmd, err := r.tx.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
	if err != nil {
		r.log.Errorw("failed to delete product", "id", productID, "error", err)
		return r.handlePgError("delete product", err)
	}
	if cmd.RowsAffected() == 0 {
		r.log.Warnw("product not found for deletion", "id", productID)
		return pkgerrors.ErrNotFound
	}
	r.log.Infow("product deleted", "id", productID)
	return nil
}

func (r *Repo) GetProductByID(ctx context.Context, productID int64) (*Product, error) {
	query := `
		SELECT id, name, description, image_url, price, stock_quantity, created_at, updated_at
		FROM products WHERE id = $1`
	row := r.tx.QueryRow(ctx, query, productID)

	var product Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.ImageUrl,
		&product.Price,
		&product.StockQuantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warnw("product not found", "id", productID)
			return nil, pkgerrors.ErrNotFound
		}
		r.log.Errorw("failed to get product", "id", productID, "error", err)
		return nil, r.handlePgError("get product by id", err)
	}
	return &product, nil
}

func (r *Repo) GetProducts(ctx context.Context) ([]*Product, error) {
	query := `
		SELECT id, name, description, image_url, price, stock_quantity, created_at, updated_at
		FROM products ORDER BY created_at DESC LIMIT 100`
	rows, err := r.tx.Query(ctx, query)
	if err != nil {
		r.log.Errorw("failed to get products", "error", err)
		return nil, r.handlePgError("get products", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID, &product.Name, &product.Description,
			&product.ImageUrl, &product.Price, &product.StockQuantity,
			&product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			r.log.Errorw("failed to scan product", "error", err)
			return nil, r.handlePgError("scan product", err)
		}
		products = append(products, &product)
	}
	if err := rows.Err(); err != nil {
		r.log.Errorw("rows iteration failed", "error", err)
		return nil, r.handlePgError("rows iteration", err)
	}
	return products, nil
}

func (r *Repo) UpdateProduct(ctx context.Context, product *Product) error {
	cmd, err := r.tx.Exec(ctx, `
		UPDATE products
		SET name = $1, description = $2, image_url = $3, price = $4, stock_quantity = $5, updated_at = NOW()
		WHERE id = $6`,
		product.Name,
		product.Description,
		product.ImageUrl,
		product.Price,
		product.StockQuantity,
		product.ID,
	)
	if err != nil {
		r.log.Errorw("failed to update product", "id", product.ID, "error", err)
		return r.handlePgError("update product", err)
	}
	if cmd.RowsAffected() == 0 {
		r.log.Warnw("product not found for update", "id", product.ID)
		return pkgerrors.ErrNotFound
	}
	r.log.Infow("product updated", "id", product.ID)
	return nil
}

func (r *Repo) handlePgError(context string, err error) error {
	if err == nil {
		return nil
	}
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pkgerrors.PGErrForeignKeyViolation,
			pkgerrors.PGErrInvalidTextRep,
			pkgerrors.PGErrInvalidType:
			return pkgerrors.ErrInvalidInput
		case pkgerrors.PGErrUniqueViolation:
			return pkgerrors.ErrAlreadyExists
		default:
			return fmt.Errorf("%s: %w", context, pkgerrors.ErrInternal)
		}
	}
	return fmt.Errorf("%s: %w", context, err)
}
