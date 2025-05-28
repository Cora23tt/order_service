package order

import (
	"context"
	"time"

	repo "github.com/Cora23tt/order_service/internal/repository/order"
	"github.com/Cora23tt/order_service/pkg/enums"
	"github.com/Cora23tt/order_service/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	repo *repo.Repo
	log  *zap.SugaredLogger
}

func NewService(r *repo.Repo, log *zap.SugaredLogger) *Service {
	return &Service{repo: r, log: log}
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
	var total int64
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
		s.log.Errorw("create order failed", "user_id", input.UserID, "error", err)
		switch err {
		case errors.ErrInvalidInput, errors.ErrAlreadyExists:
			return 0, err
		default:
			return 0, errors.ErrInternal
		}
	}

	s.log.Infow("order created", "order_id", orderID, "user_id", input.UserID)
	return orderID, nil
}

func (s *Service) GetOrderByID(ctx context.Context, orderID int64) (*repo.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		s.log.Errorw("get order by id failed", "order_id", orderID, "error", err)
		switch err {
		case errors.ErrNotFound:
			return nil, err
		default:
			return nil, errors.ErrInternal
		}
	}
	return order, nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int64) ([]*repo.Order, error) {
	orders, err := s.repo.GetAllByUser(ctx, userID)
	if err != nil {
		s.log.Errorw("get user orders failed", "user_id", userID, "error", err)
		return nil, errors.ErrInternal
	}
	return orders, nil
}

func (s *Service) CancelOrder(ctx context.Context, orderID, userID int64) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		s.log.Errorw("cancel order: get by id failed", "order_id", orderID, "error", err)
		switch err {
		case errors.ErrNotFound:
			return err
		default:
			return errors.ErrInternal
		}
	}

	if order.UserID != userID {
		s.log.Warnw("cancel order: forbidden", "order_id", orderID, "request_user_id", userID, "owner_user_id", order.UserID)
		return errors.ErrUnauthorized
	}

	if order.Status != enums.StatusPendingPayment {
		s.log.Warnw("cancel order: invalid status", "order_id", orderID, "status", order.Status)
		return errors.ErrInvalidInput
	}

	err = s.repo.UpdateStatus(ctx, orderID, "cancelled")
	if err != nil {
		s.log.Errorw("cancel order: update status failed", "order_id", orderID, "error", err)
		switch err {
		case errors.ErrNotFound, errors.ErrInvalidInput:
			return err
		default:
			return errors.ErrInternal
		}
	}

	s.log.Infow("order cancelled", "order_id", orderID)
	return nil
}

func (s *Service) AdminUpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	err := s.repo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		s.log.Errorw("admin update order status failed", "order_id", orderID, "status", status, "error", err)
		switch err {
		case errors.ErrNotFound, errors.ErrInvalidInput:
			return err
		default:
			return errors.ErrInternal
		}
	}
	s.log.Infow("admin updated order status", "order_id", orderID, "status", status)
	return nil
}

func (s *Service) DeleteOrder(ctx context.Context, orderID int64) error {
	err := s.repo.Delete(ctx, orderID)
	if err != nil {
		s.log.Errorw("delete order failed", "order_id", orderID, "error", err)
		switch err {
		case errors.ErrNotFound:
			return err
		default:
			return errors.ErrInternal
		}
	}
	s.log.Infow("order deleted", "order_id", orderID)
	return nil
}
