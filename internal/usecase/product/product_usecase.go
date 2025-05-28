package product

import (
	"context"
	"time"

	productRepo "github.com/Cora23tt/order_service/internal/repository/product"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	repo *productRepo.Repo
	log  *zap.SugaredLogger
}

func NewService(repo *productRepo.Repo, log *zap.SugaredLogger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) AddProduct(ctx context.Context, price, quantity int64, name, description, imageURL string) error {
	product := productRepo.Product{
		Name:          name,
		Price:         price,
		Description:   description,
		ImageUrl:      imageURL,
		StockQuantity: quantity,
	}
	err := s.repo.CreateProduct(ctx, &product)
	switch err {
	case nil:
		s.log.Infow("product added", "name", name, "price", price)
		return nil
	case pkgerrors.ErrInvalidInput, pkgerrors.ErrAlreadyExists:
		s.log.Warnw("invalid input for product creation", "error", err)
		return err
	default:
		s.log.Errorw("failed to create product", "error", err)
		return pkgerrors.ErrInternal
	}
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	err := s.repo.DeleteProduct(ctx, id)
	switch err {
	case nil:
		s.log.Infow("product deleted", "id", id)
		return nil
	case pkgerrors.ErrNotFound:
		s.log.Warnw("product not found for delete", "id", id)
		return err
	default:
		s.log.Errorw("failed to delete product", "id", id, "error", err)
		return pkgerrors.ErrInternal
	}
}

func (s *Service) GetProductByID(ctx context.Context, id int64) (*productRepo.Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	switch err {
	case nil:
		return product, nil
	case pkgerrors.ErrNotFound:
		s.log.Warnw("product not found", "id", id)
		return nil, err
	default:
		s.log.Errorw("failed to get product", "id", id, "error", err)
		return nil, pkgerrors.ErrInternal
	}
}

func (s *Service) GetProducts(ctx context.Context) ([]*productRepo.Product, error) {
	products, err := s.repo.GetProducts(ctx)
	if err != nil {
		s.log.Errorw("failed to get products", "error", err)
		return nil, pkgerrors.ErrInternal
	}
	return products, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id, quantity, price int64, name, description, imageUrl string) error {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		switch err {
		case pkgerrors.ErrNotFound:
			s.log.Warnw("product not found for update", "id", id)
			return err
		default:
			s.log.Errorw("failed to get product for update", "id", id, "error", err)
			return pkgerrors.ErrInternal
		}
	}

	if name != "" {
		product.Name = name
	}
	if description != "" {
		product.Description = description
	}
	if price != 0 {
		product.Price = price
	}
	if imageUrl != "" {
		product.ImageUrl = imageUrl
	}
	if quantity != 0 {
		product.StockQuantity = quantity
	}
	product.UpdatedAt = time.Now()

	err = s.repo.UpdateProduct(ctx, product)
	switch err {
	case nil:
		s.log.Infow("product updated", "id", id)
		return nil
	case pkgerrors.ErrNotFound, pkgerrors.ErrInvalidInput, pkgerrors.ErrAlreadyExists:
		s.log.Warnw("product update issue", "id", id, "error", err)
		return err
	default:
		s.log.Errorw("failed to update product", "id", id, "error", err)
		return pkgerrors.ErrInternal
	}
}
