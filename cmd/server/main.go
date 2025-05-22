package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/dig"
	"go.uber.org/zap"

	authRepo "github.com/Cora23tt/order_service/internal/repository/auth"
	authHandler "github.com/Cora23tt/order_service/internal/rest/handlers/auth"
	authService "github.com/Cora23tt/order_service/internal/usecase/auth"

	orderRepo "github.com/Cora23tt/order_service/internal/repository/order"
	orderHandler "github.com/Cora23tt/order_service/internal/rest/handlers/order"
	orderService "github.com/Cora23tt/order_service/internal/usecase/order"

	"github.com/Cora23tt/order_service/internal/rest"
	"github.com/Cora23tt/order_service/internal/rest/middleware"
	"github.com/Cora23tt/order_service/pkg/db"
)

func main() {
	var (
		port = "9999"
		host = "0.0.0.0"
		dsn  = "postgres://postgres:postgres@localhost:5432/servicedb"
	)

	if err := execute(host, port, dsn); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func execute(host, port, dsn string) error {
	deps := []any{
		func() (*pgxpool.Pool, error) {
			return db.NewDB(dsn)
		},

		gin.New,

		func() (*zap.SugaredLogger, error) {
			logger, err := zap.NewDevelopment()
			if err != nil {
				return nil, err
			}
			return logger.Sugar(), nil
		},

		middleware.New,

		authHandler.NewHandler,
		authRepo.NewRepo,
		authService.NewService,

		orderRepo.NewRepo,
		orderService.NewService,
		orderHandler.NewHandler,

		rest.NewServer,

		func(server *rest.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		if err := container.Provide(dep); err != nil {
			return err
		}
	}

	if err := container.Invoke(func(server *rest.Server) {
		server.Init()
	}); err != nil {
		return err
	}

	return container.Invoke(func(server *http.Server) error {
		return server.ListenAndServe()
	})
}
