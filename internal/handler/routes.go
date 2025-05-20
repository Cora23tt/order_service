package handler

import (
	"github.com/Cora23tt/order_service/internal/handler/auth"
	"github.com/Cora23tt/order_service/internal/handler/order"
	"github.com/Cora23tt/order_service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func InitRoutes(authH *auth.Handler, orderH *order.Handler, authMW *middleware.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authH.Register)
		authGroup.POST("/login", authH.Login)
	}

	userOrders := api.Group("/orders", authMW.Authorize("user"))
	{
		userOrders.POST("/", orderH.Create)
		userOrders.GET("/", orderH.ListOwn)
		userOrders.GET("/:id", orderH.GetOwn)
		userOrders.PUT("/:id", orderH.Update)
		userOrders.PUT("/:id/cancel", orderH.Cancel)
	}

	adminOrders := api.Group("/admin/orders", authMW.Authorize("admin"))
	{
		adminOrders.GET("/", orderH.ListAll)
		adminOrders.PUT("/:id/status", orderH.UpdateStatus)
		adminOrders.DELETE("/:id", orderH.Delete)
		adminOrders.GET("/stats", orderH.Stats)
		adminOrders.GET("/export/json", orderH.ExportJSON)
		adminOrders.GET("/export/csv", orderH.ExportCSV)
		adminOrders.POST("/simulate/traffic", orderH.Simulate)
	}

	return r
}
