package service

import (
	"context"
	"time"

	"e-commerce.com/internal/models"
	"e-commerce.com/internal/repository"
	"github.com/google/uuid"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *models.CreateProductRequest, sellerId string) error
	GetSellerProducts(ctx context.Context, sellerId string, page int, limit int) ([]*models.Product, int64, bool, error)
	UpdateProduct(ctx context.Context, productId string, product *models.UpdateProductRequest) error
	DeleteProduct(ctx context.Context, productId string) error
	GetProductById(ctx context.Context, productId string) (*models.Product, []models.ProductReview, error)
	GetAllProducts(ctx context.Context, search *string, limit, offset int) (*models.ProductResponse, error)
}

type productService struct {
	repo repository.ProductRepo
}

func (s *productService) GetProductById(ctx context.Context, productId string) (*models.Product, []models.ProductReview, error) {
	return s.repo.GetProductByID(ctx, productId)

}

func (s *productService) UpdateProduct(ctx context.Context, productId string, product *models.UpdateProductRequest) error {
	return s.repo.UpdateProduct(ctx, productId, product)
}

func (s *productService) DeleteProduct(ctx context.Context, productId string) error {
	return s.repo.DeleteProduct(ctx, productId)
}

func (s *productService) GetSellerProducts(ctx context.Context, sellerId string, page int, limit int) ([]*models.Product, int64, bool, error) {
	return s.repo.GetSellerProducts(ctx, sellerId, page, limit)
}

func (s *productService) CreateProduct(ctx context.Context, product *models.CreateProductRequest, sellerId string) error {
	productModel := &models.Product{
		ID:          uuid.New().String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    product.Quantity,
		Discount:    product.Discount,
		SellerID:    sellerId,
		Category:    product.Category,
		Images:      product.Images,
		Stock:       product.Stock,
		Rating:      0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err := s.repo.CreateProduct(ctx, productModel)
	if err != nil {
		return err
	}
	return nil
}

func (s *productService) GetAllProducts(ctx context.Context, search *string, limit, offset int) (*models.ProductResponse, error) {
	return s.repo.GetAllProducts(ctx, search, limit, offset)
}

func NewProductService(repo repository.ProductRepo) ProductService {
	return &productService{repo: repo}
}
