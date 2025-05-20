package main

import (
	"log"
	"net/http"

	"go.uber.org/dig"

	"github.com/Cora23tt/order_service/internal/config"
	"github.com/Cora23tt/order_service/internal/handler"
	"github.com/Cora23tt/order_service/internal/middleware"
	"github.com/Cora23tt/order_service/pkg/db"

	handlerOrder "github.com/Cora23tt/order_service/internal/handler/order"
	repo "github.com/Cora23tt/order_service/internal/repo/postgres"
	service "github.com/Cora23tt/order_service/internal/service/order"
)

func main() {
	c := dig.New()

	c.Provide(config.Load)
	c.Provide(db.NewDB)

	c.Provide(repo.NewRepository)
	c.Provide(service.NewService)
	c.Provide(handlerOrder.NewHandler)

	c.Provide(middleware.NewAuthMiddleware)
	c.Provide(handler.InitRoutes)

	err := c.Invoke(func(cfg *config.Config, router http.Handler) {
		log.Println("server started at :" + cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
	})
	if err != nil {
		log.Fatal(err)
	}
}
