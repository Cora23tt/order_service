package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	repo "github.com/Cora23tt/order_service/internal/repository/order"
	"github.com/Cora23tt/order_service/pkg/enums"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
)

type Service struct {
	repo *repo.Repo
}

func NewService(r *repo.Repo) *Service {
	return &Service{repo: r}
}

type CreateOrderInput struct {
	UserID       int64
	Items        []OrderItemInput
	PickupPoint  string
	DeliveryDate *time.Time
}

type OrderItemInput struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int64 `json:"quantity" binding:"required"`
	Price     int64 `json:"price" binding:"required"`
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (int64, error) {
	var total int64 = 0
	for _, item := range input.Items {
		total += item.Price * item.Quantity
	}

	order := &repo.Order{
		UserID:       input.UserID,
		Status:       enums.StatusPendingPayment,
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

func (s *Service) GetOrderByID(ctx context.Context, orderID int64) (*repo.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if err == pkgerrors.ErrNotFound {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	return order, nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int64) ([]*repo.Order, error) {
	return s.repo.GetAllByUser(ctx, userID)
}

func (s *Service) AdminUpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	return s.repo.UpdateStatus(ctx, orderID, status)
}

func (s *Service) DeleteOrder(ctx context.Context, orderID int64) error {
	return s.repo.Delete(ctx, orderID)
}

func (s *Service) CancelOrder(ctx context.Context, orderID, userID int64) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			return pkgerrors.ErrNotFound
		}
		return fmt.Errorf("get order by id: %w", err)
	}

	if order.UserID != userID {
		return pkgerrors.ErrUnauthorized
	}

	if order.Status != enums.StatusPendingPayment {
		return pkgerrors.ErrInvalidInput // нельзя отменить заказ
	}

	return s.repo.UpdateStatus(ctx, orderID, "cancelled")
}
