package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Cora23tt/order_service/internal/repository/auth"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/dgrijalva/jwt-go/v4"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = time.Duration(12 * time.Hour)

type Service struct {
	repo *auth.Repo
}

type User struct {
	CreatedAt      time.Time
	PhoneNumber    string
	HashedPassword string
	ID             int
}

func NewService(r *auth.Repo) *Service {
	return &Service{repo: r}
}

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

func (s *Service) CreateUser(ctx context.Context, phoneNumber, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	user := auth.User{
		PhoneNumber:    phoneNumber,
		HashedPassword: string(hashedPassword),
	}
	userID, err := s.repo.Create(ctx, &user)

	if errors.Is(err, pkgerrors.ErrAlreadyExists) {
		return 0, pkgerrors.ErrAlreadyExists
	}
	return userID, err
}

func (s *Service) Validate(ctx context.Context, username, password string) (bool, error) {
	user, err := s.repo.GetUser(ctx, username)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil, nil
}

func (s *Service) GetUserID(ctx context.Context, token string) (string, error) {
	return s.repo.GetUserID(ctx, token)
}

func (s *Service) GenerateJWT(ctx context.Context, phoneNumber, password string) (string, error) {
	user, err := s.repo.GetUser(ctx, phoneNumber)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			return "", pkgerrors.ErrNotFound
		} else {
			return "", pkgerrors.ErrInternal
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", pkgerrors.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		tokenClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: &jwt.Time{Time: time.Now().Add(tokenTTL)},
				IssuedAt:  &jwt.Time{Time: time.Now()},
			},
			UserId: user.ID,
		},
	)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
