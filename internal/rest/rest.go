package rest

import (
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Cora23tt/order_service/docs"
	"github.com/Cora23tt/order_service/internal/rest/handlers/auth"
	"github.com/Cora23tt/order_service/internal/rest/handlers/order"
	"github.com/Cora23tt/order_service/internal/rest/handlers/product"
	"github.com/Cora23tt/order_service/internal/rest/handlers/user"
	"github.com/Cora23tt/order_service/internal/rest/middleware"
)

type Server struct {
	mux        *gin.Engine
	auth       *auth.Handler
	order      *order.Handler
	product    *product.Handler
	user       *user.Handler
	middleware *middleware.Middleware
}

func NewRESTServer(
	mux *gin.Engine,
	auth *auth.Handler,
	order *order.Handler,
	mdlwr *middleware.Middleware,
	product *product.Handler,
	user *user.Handler,
) *Server {
	return &Server{
		mux:        mux,
		auth:       auth,
		order:      order,
		product:    product,
		user:       user,
		middleware: mdlwr,
	}
}

func NewHTTPServer(s *Server) *http.Server {
	return &http.Server{
		Addr:    net.JoinHostPort(os.Getenv("HOST"), os.Getenv("PORT")),
		Handler: s.mux,
	}
}

func (s *Server) SetupRoutes() {
	const baseUrl = "/api/v1"

	s.mux.Use(gin.Recovery())
	s.mux.Use(s.middleware.ZapLogger())
	s.mux.Use(s.middleware.CORSMiddleware())

	s.mux.GET("/profile/:id/photo", s.user.GetProfilePhoto)
	s.mux.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authGroup := s.mux.Group(baseUrl + "/auth")
	{
		authGroup.POST("/signin", s.auth.SignIn)
		authGroup.POST("/signup", s.auth.SignUp)
	}

	userGroup := s.mux.Group(baseUrl+"/me", s.middleware.AuthWithRoles("user", "admin"))
	{
		userGroup.GET("/", s.user.GetProfile)
		userGroup.PATCH("/", s.user.UpdateProfile)
	}
	adminUserGroup := s.mux.Group(baseUrl+"/admin/users", s.middleware.AuthWithRoles("admin"))
	{
		adminUserGroup.GET("/", s.user.ListUsers)
	}

	ordersGroup := s.mux.Group(baseUrl+"/orders", s.middleware.AuthWithRoles("user", "admin"))
	{
		ordersGroup.POST("/", s.order.Create)
		ordersGroup.GET("/", s.order.GetAll)
		ordersGroup.GET("/:id", s.order.GetByID)
		ordersGroup.GET("/:id/cancel", s.order.Cancel)
	}
	adminOrdersGroup := s.mux.Group(baseUrl+"/orders", s.middleware.AuthWithRoles("admin"))
	{
		adminOrdersGroup.DELETE("/:id", s.order.Delete)
		adminOrdersGroup.PUT("/:id", s.order.Update)
		adminOrdersGroup.GET("/stats", s.order.GetStats)
		adminOrdersGroup.GET("/export", s.order.Export)
		adminOrdersGroup.GET("/export/csv", s.order.ExportCSV)
	}

	publicProductGroup := s.mux.Group(baseUrl + "/products")
	{
		publicProductGroup.GET("/", s.product.GetProducts)
		publicProductGroup.GET("/:id", s.product.GetProduct)
	}
	adminProductGroup := s.mux.Group(baseUrl+"/products", s.middleware.AuthWithRoles("admin"))
	{
		adminProductGroup.POST("/", s.product.AddProduct)
		adminProductGroup.PUT("/:id", s.product.UpdateProduct)
		adminProductGroup.DELETE("/:id", s.product.DeleteProduct)
	}
}
