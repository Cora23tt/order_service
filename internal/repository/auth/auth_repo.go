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
	CreatedAt      time.Time
	PhoneNumber    string
	HashedPassword string
	ID             int
}

func (r *Repo) Create(ctx context.Context, user *User) (int, error) {
	userID := 0
	err := r.db.QueryRow(ctx, `
        INSERT INTO users (phone_number, password_hash)
        VALUES ($1, $2)
        RETURNING id
    `, user.PhoneNumber, user.HashedPassword).Scan(&userID)
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == "23505" {
			return 0, pkgErrors.ErrAlreadyExists
		}
		return 0, err
	}
	return userID, nil
}

func (r *Repo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id, phone_number, password_hash, created_at
		FROM users
		WHERE phone_number = $1
	`, phoneNumber).Scan(&user.ID, &user.PhoneNumber, &user.HashedPassword, &user.CreatedAt)
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

func (r *Repo) GetUser(ctx context.Context, phoneNumber string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id, phone_number, password_hash
		FROM users
		WHERE phone_number = $1
	`, phoneNumber).Scan(&user.ID, &user.PhoneNumber, &user.HashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, pkgErrors.ErrNotFound
		}
		return User{}, err
	}

	return user, nil
}
