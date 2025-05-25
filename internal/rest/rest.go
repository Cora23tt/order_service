package rest

import (
	"net"
	"net/http"
	"os"

	"github.com/Cora23tt/order_service/internal/rest/handlers/product"

	"github.com/gin-gonic/gin"

	"github.com/Cora23tt/order_service/internal/rest/handlers/auth"
	"github.com/Cora23tt/order_service/internal/rest/handlers/order"
	"github.com/Cora23tt/order_service/internal/rest/middleware"
)

type Server struct {
	mux        *gin.Engine
	order      *order.Handler
	auth       *auth.Handler
	product    *product.Handler
	middleware *middleware.Middleware
}

func NewRESTServer(mux *gin.Engine, auth *auth.Handler,
	order *order.Handler, mdlwr *middleware.Middleware,
	product *product.Handler) *Server {
	return &Server{
		mux:        mux,
		order:      order,
		auth:       auth,
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

	publicGroup := s.mux.Group(baseUrl + "/public")
	{
		publicGroup.POST("/signin", s.auth.SignIn)
		publicGroup.POST("/signup", s.auth.SignUp)
	}

	secureGroup := s.mux.Group(baseUrl + "/secure").Use(s.middleware.Auth())
	{
		secureGroup.POST("/orders", s.order.Create)
		secureGroup.GET("/orders", s.order.GetAll)
		secureGroup.GET("/orders/:id", s.order.GetByID)
		secureGroup.PUT("/orders/:id", s.order.Update)
		secureGroup.DELETE("/orders/:id", s.order.Delete)
	}
	productGroup := s.mux.Group(baseUrl + "/products")
	{
		productGroup.POST("/", s.product.AddProduct)
		productGroup.GET("/", s.product.GetProducts)
		productGroup.GET("/:id", s.product.GetProduct)
		productGroup.DELETE("/:id", s.product.DeleteProduct)
		productGroup.PUT("/:id", s.product.UpdateProduct)
	}
}
