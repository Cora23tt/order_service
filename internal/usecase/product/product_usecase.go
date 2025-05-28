package product

import (
	"context"
	"time"

	productRepo "github.com/Cora23tt/order_service/internal/repository/product"
)

type Service struct {
	repo *productRepo.Repo
}

func NewService(repo *productRepo.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddProduct(ctx context.Context, price int64, quantity int64, name, description, imageURL string) error {
	product := productRepo.Product{
		Name:          name,
		Price:         price,
		Description:   description,
		ImageUrl:      imageURL,
		StockQuantity: quantity,
	}
	return s.repo.CreateProduct(ctx, &product)
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) error {
	return s.repo.DeleteProduct(ctx, id)
}

func (s *Service) GetProductByID(ctx context.Context, id int64) (*productRepo.Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *Service) GetProducts(ctx context.Context) ([]*productRepo.Product, error) {
	return s.repo.GetProducts(ctx)
}
func (s *Service) UpdateProduct(ctx context.Context, id int64, quantity int64, price int64, name, description, imageUrl string) error {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return err
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
	return s.repo.UpdateProduct(ctx, product)
}
