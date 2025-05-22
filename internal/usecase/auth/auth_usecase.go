package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Cora23tt/order_service/internal/repository/auth"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/Cora23tt/order_service/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *auth.Repo
}

type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewService(r *auth.Repo) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateUser(ctx context.Context, username, hashedPassword string) (string, error) {
	user := auth.User{
		ID:             utils.NewUUID(),
		Username:       username,
		HashedPassword: hashedPassword,
	}
	err := s.repo.Create(ctx, &user)

	if errors.Is(err, pkgerrors.ErrAlreadyExists) {
		return "", pkgerrors.ErrAlreadyExists
	}
	return user.ID, err
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:             u.ID,
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
		CreatedAt:      u.CreatedAt,
	}, nil
}

func (s *Service) Validate(ctx context.Context, username, password string) (bool, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil, nil
}

func (s *Service) GetUserID(ctx context.Context, token string) (string, error) {
	return s.repo.GetUserID(ctx, token)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:             user.ID,
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		CreatedAt:      user.CreatedAt,
	}, nil
}

func (s *Service) CreateSessionToken(ctx context.Context, userID string) (string, error) {
	token := utils.NewUUID()
	err := s.repo.StoreSessionToken(ctx, token, userID)
	if err != nil {
		return "", err
	}
	return token, nil
}
