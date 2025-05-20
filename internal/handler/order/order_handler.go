package order

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/orders")
	{
		grp.POST("/", h.Create)
		grp.GET("/", h.List)
	}
}

func (h *Handler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "create order"})
}

func (h *Handler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "list orders"})
}
