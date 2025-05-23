package product

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Product struct {
	ID            int
	Name          string
	Description   string
	ImageUrl      string
	Price         int64
	StockQuantity int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateProduct(ctx context.Context, product *Product) error {
	query := `INSERT INTO products(name,description,image_url,price,stock_quantity)
			  VALUES ($1,$2,$3,$4,$5)`
	_, err := r.db.Exec(ctx, query, product.Name, product.Description, product.ImageUrl, product.Price, product.StockQuantity)
	if err != nil {
		return err
	}
	return nil
}
func (r *Repo) DeleteProduct(ctx context.Context, productID int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(ctx, query, productID)
	if err != nil {
		return err
	}
	return nil
}
func (r *Repo) GetProductByID(ctx context.Context, productID int) (*Product, error) {
	query := `SELECT id,name,description,image_url,price,stock_quantity,created_at,updated_at
			  FROM products
			  WHERE id = $1`
	row := r.db.QueryRow(ctx, query, productID)
	product := Product{}
	err := row.Scan(product.ID, &product.Name, &product.Description, &product.ImageUrl, &product.Price, &product.StockQuantity, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &Product{}, ErrProductNotFound
		}
		return &Product{}, err
	}
	return &product, nil
}
func (r *Repo) GetProducts(ctx context.Context) ([]*Product, error) {
	query := `SELECT id,name,description,image_url,price,stock_quantity,created_at,updated_at`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	var products []*Product
	for rows.Next() {
		product := Product{}
		err := rows.Scan(product.ID, &product.Name, &product.Description, &product.ImageUrl, &product.Price, &product.StockQuantity, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	return products, nil
}
func (r *Repo) UpdateProduct(ctx context.Context, product *Product) error {
	query := `UPDATE products 
              SET name = $1,description = $2,image_url = $3,price = $4,stock_quantity = $5
              WHERE id = $6`
	_, err := r.db.Exec(ctx, query, product.Name, product.Description, product.ImageUrl, product.Price, product.StockQuantity, product.ID)
	if err != nil {
		return err
	}
	return nil
}

var (
	ErrProductNotFound = errors.New("product not found")
)
