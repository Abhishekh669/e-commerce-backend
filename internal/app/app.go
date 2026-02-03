// internal/app/app.go
package app

import (
	"e-commerce.com/internal/db"
	"e-commerce.com/internal/handler"
	"e-commerce.com/internal/repository"
	"e-commerce.com/internal/service"
)

type App struct {
	// Repositories
	UserRepo repository.UserRepo
	// Add other repos as needed

	// Services
	UserService service.UserService
	// Add other services as needed

	// Handlers
	UserHandler *handler.UserHandler
	// Add other handlers as needed

	ProductHandler *handler.ProductHandler
	ProductService service.ProductService
	ProductRepo    repository.ProductRepo

	PaymentHandler *handler.PaymentHandler
	PaymentService service.PaymentService
	PaymentRepo    repository.PaymentRepo

	OrderHandler *handler.OrderHandler
	OrderService service.OrderService
	OrderRepo    repository.OrderRepo

	CommentHandler *handler.CommentHandler
	CommentService service.CommentService
	CommentRepo    repository.CommentRepo
}

func New() (*App, error) {

	userRepo := repository.NewUserRepository()
	productRepo := repository.NewProductRepository()
	paymentRepo := repository.NewPaymentRepository()
	orderRepo := repository.NewOrderRepository()
	commentRepo := repository.NewCommentRepositry()

	// Initialize services
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	paymentService := service.NewPaymentService(paymentRepo)
	orderService := service.NewOrderService(orderRepo, productRepo)
	commentService := service.NewCommnetService(commentRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	orderHandler := handler.NewOrderHandler(orderService)
	comentHandler := handler.NewCommentHandler(commentService)

	return &App{
		UserRepo:       userRepo,
		UserService:    userService,
		UserHandler:    userHandler,
		ProductHandler: productHandler,
		ProductService: productService,
		ProductRepo:    productRepo,
		PaymentHandler: paymentHandler,
		PaymentService: paymentService,
		PaymentRepo:    paymentRepo,
		OrderHandler:   orderHandler,
		OrderService:   orderService,
		OrderRepo:      orderRepo,
		CommentRepo:    commentRepo,
		CommentHandler: comentHandler,
		CommentService: commentService,
	}, nil
}

func (a *App) Close() {
	db.Cleanup()
}
