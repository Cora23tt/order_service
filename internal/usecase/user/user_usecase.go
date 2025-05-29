package user

import (
	"context"
	"fmt"

	"github.com/Cora23tt/order_service/internal/repository/user"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
)

type Service struct {
	repo *user.Repo
}

func NewService(r *user.Repo) *Service {
	return &Service{repo: r}
}

type Profile struct {
	ID          int64   `json:"id" example:"1"`
	PhoneNumber string  `json:"phone_number" example:"+998901234567"`
	Role        string  `json:"role" example:"user"`
	PINFL       *string `json:"pinfl,omitempty" example:"12345678901234"`
	AvatarURL   *string `json:"avatar_url,omitempty" example:"/profile/1/photo"`
}

func (s *Service) GetProfile(ctx context.Context, userID int64) (*Profile, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}

	return &Profile{
		ID:          u.ID,
		PhoneNumber: u.PhoneNumber,
		Role:        u.Role,
		PINFL:       u.PINFL,
		AvatarURL:   u.AvatarURL,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, pinfl *string, avatarURL *string) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			return pkgerrors.ErrNotFound
		}
		return fmt.Errorf("get user before update: %w", err)
	}

	u.PINFL = pinfl
	u.AvatarURL = avatarURL

	if err := s.repo.Update(ctx, u); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	return nil
}

func (s *Service) ListAllUsers(ctx context.Context) ([]Profile, error) {
	users, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	var result []Profile
	for _, u := range users {
		result = append(result, Profile{
			ID:          u.ID,
			PhoneNumber: u.PhoneNumber,
			Role:        u.Role,
			PINFL:       u.PINFL,
			AvatarURL:   u.AvatarURL,
		})
	}

	return result, nil
}
