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
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

func (s *Service) CreateUser(ctx context.Context, phoneNumber, password string) (int64, error) {
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

func (s *Service) Validate(token string) (int64, string, error) {
	return s.ParseToken(token)
}

func (s *Service) GetUserID(ctx context.Context, token string) (int64, error) {
	return s.repo.GetUserID(ctx, token)
}

func (s *Service) GenerateJWT(ctx context.Context, phoneNumber, password string) (string, error) {
	user, err := s.repo.GetUser(ctx, phoneNumber)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			return "", pkgerrors.ErrNotFound
		}
		return "", pkgerrors.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return "", pkgerrors.ErrInvalidCredentials
	}

	claims := tokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: &jwt.Time{Time: time.Now().Add(tokenTTL)},
			IssuedAt:  &jwt.Time{Time: time.Now()},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	return token.SignedString([]byte(secret))
}

func (s *Service) ParseToken(tokenStr string) (int64, string, error) {
	claims := &tokenClaims{}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, "", fmt.Errorf("JWT_SECRET not set")
	}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return 0, "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, claims.Role, nil
}
