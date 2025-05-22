package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitAllSchemas(ctx context.Context, db *pgxpool.Pool) error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS notes (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT CHECK (
				status IN ('pending', 'in_progress', 'completed')
			) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`CREATE TABLE IF NOT EXISTS events (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			location TEXT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			status TEXT NOT NULL CHECK (
				status IN ('scheduled', 'ongoing', 'finished', 'cancelled')
			)
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			hashed_password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS sessions (
			token UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
