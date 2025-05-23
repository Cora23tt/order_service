package order

import (
	"github.com/Cora23tt/order_service/internal/usecase/order"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *order.Service
}

func NewHandler(service *order.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {

	return
}

func (h *Handler) GetAll(c *gin.Context) {
	return
}

func (h *Handler) GetByID(c *gin.Context) {
	return
}

func (h *Handler) Update(c *gin.Context) {
	return
}

func (h *Handler) Delete(c *gin.Context) {
	return
}
