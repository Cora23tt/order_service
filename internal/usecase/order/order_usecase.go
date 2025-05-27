package order

import (
	"context"
	"fmt"
	"time"

	repo "github.com/Cora23tt/order_service/internal/repository/order"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
)

type Service struct {
	repo *repo.Repo
}

func NewService(r *repo.Repo) *Service {
	return &Service{repo: r}
}

type CreateOrderInput struct {
	UserID       int
	Items        []OrderItemInput
	PickupPoint  string
	DeliveryDate *time.Time
}

type OrderItemInput struct {
	ProductID int
	Quantity  int
	Price     int
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (int, error) {
	total := 0
	for _, item := range input.Items {
		total += item.Price * item.Quantity
	}

	order := &repo.Order{
		UserID:       input.UserID,
		Status:       "pending_payment",
		DeliveryDate: input.DeliveryDate,
		PickupPoint:  input.PickupPoint,
		TotalAmount:  total,
		Items:        make([]repo.OrderItem, 0, len(input.Items)),
	}

	for _, item := range input.Items {
		order.Items = append(order.Items, repo.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	orderID, err := s.repo.Create(ctx, order)
	if err != nil {
		if err == pkgerrors.ErrInvalidInput {
			return 0, pkgerrors.ErrInvalidInput
		}
		return 0, pkgerrors.ErrInternal
	}

	return orderID, nil
}

func (s *Service) GetOrderByID(ctx context.Context, orderID int) (*repo.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return order, nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int) ([]*repo.Order, error) {
	return s.repo.GetAllByUser(ctx, userID)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	return s.repo.UpdateStatus(ctx, orderID, status)
}

func (s *Service) DeleteOrder(ctx context.Context, orderID int) error {
	return s.repo.Delete(ctx, orderID)
}
