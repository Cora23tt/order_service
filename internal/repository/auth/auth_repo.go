package auth

import (
	"context"
	"time"

	"errors"

	pkgErrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type User struct {
	ID             string
	Username       string
	HashedPassword string
	CreatedAt      time.Time
}

func (r *Repo) Create(ctx context.Context, user *User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, username, hashed_password)
		VALUES ($1, $2, $3)
	`, user.ID, user.Username, user.HashedPassword)
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == "23505" {
			return pkgErrors.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *Repo) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, hashed_password, created_at
		FROM users
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.HashedPassword, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgErrors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repo) GetUserID(ctx context.Context, token string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, `
		SELECT user_id
		FROM sessions
		WHERE token = $1
	`, token).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", pkgErrors.ErrNotFound
		}
		return "", err
	}
	return userID, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, hashed_password, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.HashedPassword, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgErrors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repo) StoreSessionToken(ctx context.Context, token, userID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO sessions (token, user_id)
		VALUES ($1, $2)
	`, token, userID)
	if err != nil {
		return err
	}
	return nil
}
