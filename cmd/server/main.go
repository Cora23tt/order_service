package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/dig"

	authRepo "github.com/Cora23tt/order_service/internal/repository/auth"
	authHandler "github.com/Cora23tt/order_service/internal/rest/handlers/auth"
	authService "github.com/Cora23tt/order_service/internal/usecase/auth"

	userRepo "github.com/Cora23tt/order_service/internal/repository/user"
	userHandler "github.com/Cora23tt/order_service/internal/rest/handlers/user"
	userService "github.com/Cora23tt/order_service/internal/usecase/user"

	orderRepo "github.com/Cora23tt/order_service/internal/repository/order"
	orderHandler "github.com/Cora23tt/order_service/internal/rest/handlers/order"
	orderService "github.com/Cora23tt/order_service/internal/usecase/order"

	productRepo "github.com/Cora23tt/order_service/internal/repository/product"
	productHandler "github.com/Cora23tt/order_service/internal/rest/handlers/product"
	productService "github.com/Cora23tt/order_service/internal/usecase/product"

	"github.com/Cora23tt/order_service/internal/rest"
	"github.com/Cora23tt/order_service/internal/rest/middleware"
	"github.com/Cora23tt/order_service/pkg/db"
	"github.com/Cora23tt/order_service/pkg/logger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default environment variables:", err)
	}
	if err := execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func execute() error {
	deps := []any{
		logger.New,
		db.NewDB,
		gin.New,
		func(s *authService.Service) middleware.AuthValidator { return s },
		middleware.New,

		authHandler.NewHandler,
		authRepo.NewRepo,
		authService.NewService,

		userRepo.NewRepo,
		userService.NewService,
		userHandler.NewHandler,

		orderRepo.NewRepo,
		orderService.NewService,
		orderHandler.NewHandler,

		productRepo.NewRepo,
		productHandler.NewHandler,
		productService.NewService,

		rest.NewRESTServer,
		rest.NewHTTPServer,
	}

	container := dig.New()
	for _, dep := range deps {
		if err := container.Provide(dep); err != nil {
			return fmt.Errorf("failed to provide dependency: %w", err)
		}
	}

	err := container.Invoke(
		func(server *rest.Server) {
			server.SetupRoutes()
		})
	if err != nil {
		return fmt.Errorf("failed to initialize routes: %w", err)
	}

	return container.Invoke(
		func(server *http.Server) error {
			return fmt.Errorf("failed to start server: %w", server.ListenAndServe())
		})
}
