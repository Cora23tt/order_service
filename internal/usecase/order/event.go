package order

import (
	"context"

	"github.com/Cora23tt/order_service/internal/repository/order"
)

type Service struct {
	repo *order.Repo
}

func NewService(r *order.Repo) *Service {
	return &Service{repo: r}
}

type Order struct {
}

func (s *Service) Create(ctx context.Context, e *Order) error {
	return s.repo.Create(ctx, &order.Order{})
}

func (s *Service) GetByID(ctx context.Context, id string) (*Order, error) {
	return &Order{}, nil
}

func (s *Service) GetAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	return orders, nil
}

func (s *Service) Update(ctx context.Context, e *Order) error {
	return s.repo.Update(ctx, &order.Order{})
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
