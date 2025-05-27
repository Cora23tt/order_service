package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitAllSchemas(ctx context.Context, db *pgxpool.Pool) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			role VARCHAR(16) NOT NULL DEFAULT 'user',
			phone_number VARCHAR(20) NOT NULL UNIQUE,
			pinfl VARCHAR(14),
			password_hash TEXT NOT NULL,
			avatar_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		INSERT INTO users (role, phone_number, password_hash)
		VALUES ('admin', 'admin123', '$2a$10$adKovmcwFocIAq1ut4IlXuJA83pvaCQyMiiSYwwDesAiLPD4lXxce')
		ON CONFLICT (phone_number) DO NOTHING;
		`,
		`
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			image_url TEXT,
			price INTEGER NOT NULL CHECK (price >= 0),
			stock_quantity INTEGER NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		`,
		`
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			status VARCHAR(32) NOT NULL,
			delivery_date DATE,
			pickup_point VARCHAR(255),
			order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			total_amount INTEGER NOT NULL,
			receipt_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id),
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			price INTEGER NOT NULL,
			total_price INTEGER GENERATED ALWAYS AS (quantity * price) STORED
		);
		`,
	}

	for _, q := range queries {
		if _, err := db.Exec(ctx, q); err != nil {
			return fmt.Errorf("failed to execute schema init: %w", err)
		}
	}

	return nil
}
