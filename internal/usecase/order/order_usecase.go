package order

import (
	"context"
	"time"

	repo "github.com/Cora23tt/order_service/internal/repository/order"
	"github.com/Cora23tt/order_service/internal/repository/product"
	"github.com/Cora23tt/order_service/internal/repository/uow"
	"github.com/Cora23tt/order_service/pkg/enums"
	"github.com/Cora23tt/order_service/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	repo *repo.Repo
	log  *zap.SugaredLogger
	uow  uow.UnitOfWork
}

func NewService(r *repo.Repo, log *zap.SugaredLogger, uow uow.UnitOfWork) *Service {
	return &Service{repo: r, log: log, uow: uow}
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
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Errorw("begin transaction failed", "error", err)
		return 0, errors.ErrInternal
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	orderRepo := repo.NewWithTx(tx.GetTx(), s.log)
	productRepo := product.NewWithTx(tx.GetTx(), s.log)

	var total int64
	order := &repo.Order{
		UserID:       input.UserID,
		Status:       enums.StatusPendingPayment,
		DeliveryDate: input.DeliveryDate,
		PickupPoint:  input.PickupPoint,
		TotalAmount:  0,
		Items:        make([]repo.OrderItem, 0, len(input.Items)),
	}

	for _, item := range input.Items {
		product, err := productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			s.log.Errorw("product not found", "product_id", item.ProductID, "error", err)
			return 0, errors.ErrInvalidInput
		}
		if product.StockQuantity < item.Quantity {
			s.log.Warnw("insufficient stock", "product_id", item.ProductID, "available", product.StockQuantity, "requested", item.Quantity)
			return 0, errors.ErrInsufficientStock
		}

		total += item.Price * item.Quantity
		order.Items = append(order.Items, repo.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}
	order.TotalAmount = total

	orderID, err := orderRepo.Create(ctx, order)
	if err != nil {
		s.log.Errorw("create order failed", "user_id", input.UserID, "error", err)
		switch err {
		case errors.ErrInvalidInput, errors.ErrAlreadyExists:
			return 0, err
		default:
			return 0, errors.ErrInternal
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Errorw("commit transaction failed", "order_id", orderID, "error", err)
		return 0, errors.ErrInternal
	}
	committed = true

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
