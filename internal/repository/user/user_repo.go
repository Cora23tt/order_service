package user

import (
	"context"
	"time"

	"github.com/Cora23tt/order_service/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type User struct {
	ID           int64
	PhoneNumber  string
	PINFL        *string
	Role         string
	AvatarURL    *string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, phone_number, pinfl, role, avatar_url, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.PhoneNumber, &u.PINFL, &u.Role, &u.AvatarURL, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repo) Update(ctx context.Context, u *User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET pinfl = $1, avatar_url = $2, updated_at = NOW()
		WHERE id = $3
	`, u.PINFL, u.AvatarURL, u.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) ListAll(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, phone_number, pinfl, role, avatar_url, password_hash, created_at, updated_at
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.PhoneNumber, &u.PINFL, &u.Role, &u.AvatarURL, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
