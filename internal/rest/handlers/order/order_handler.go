package order

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Cora23tt/order_service/internal/usecase/order"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *order.Service
	log     *zap.SugaredLogger
}

func NewHandler(service *order.Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

type CreateOrderRequest struct {
	Items        []order.OrderItemInput `json:"items" binding:"required"`
	PickupPoint  string                 `json:"pickup_point" binding:"required"`
	DeliveryDate *time.Time             `json:"delivery_date,omitempty"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetInt("userID")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnw("invalid create order", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	id, err := h.service.CreateOrder(c.Request.Context(), order.CreateOrderInput{
		UserID:       userID,
		Items:        req.Items,
		PickupPoint:  req.PickupPoint,
		DeliveryDate: req.DeliveryDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrInvalidInput):
			h.log.Warnw("invalid input for order creation", "userID", userID, "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user or product"})
		default:
			h.log.Errorw("failed to create order", "userID", userID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order_id": id})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		h.log.Errorw("get order failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) GetAll(c *gin.Context) {
	userID := c.GetInt("userID")

	orders, err := h.service.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		h.log.Errorw("get orders failed", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.service.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		h.log.Errorw("update order failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.service.DeleteOrder(c.Request.Context(), id); err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		h.log.Errorw("delete order failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}
