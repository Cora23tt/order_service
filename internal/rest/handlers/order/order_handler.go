package order

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Cora23tt/order_service/internal/usecase/order"
	"github.com/Cora23tt/order_service/pkg/enums"
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
	PickupPoint  string                 `json:"pickup_point" binding:"required" example:"Mirzo Ulug'bek. Buyuk Ipak Yoli st. 109. 45"`
	DeliveryDate *time.Time             `json:"-"`
}

// Create godoc
// @Summary Create order
// @Description Create a new order (user/admin)
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order info"
// @Success 201 {object} map[string]int64 "order_id"
// @Failure 400 {object} map[string]string "invalid input"
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 409 {object} map[string]string "insufficient stock"
// @Failure 500 {object} map[string]string "internal error"
// @Router /api/v1/orders/ [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(int64)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnw("invalid create order request", "userID", userID, "error", err)
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
			h.log.Warnw("invalid data for order creation", "userID", userID, "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user or product"})
		case errors.Is(err, pkgerrors.ErrInsufficientStock):
			h.log.Warnw("insufficient stock for order", "userID", userID, "error", err)
			c.JSON(http.StatusConflict, gin.H{"error": "insufficient stock"})
		default:
			h.log.Errorw("internal error during order creation", "userID", userID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order_id": id})
}

// @Summary Get order by ID (user/admin)
// @Description Возвращает заказ по ID
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} order.Order
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid order id", "param", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	userIDRaw, ok1 := c.Get("userID")
	roleRaw, ok2 := c.Get("role")
	if !ok1 || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(int64)
	role := roleRaw.(string)

	order, err := h.service.GetOrderByID(c.Request.Context(), id, userID, role)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case err != nil:
		h.log.Errorw("get order by id failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, order)
	}
}

// @Summary Get all user orders (user/admin)
// @Description Возвращает список заказов
// @Tags orders
// @Security BearerAuth
// @Success 200 {array} order.Order
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [get]
func (h *Handler) GetAll(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(int64)

	orders, err := h.service.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		h.log.Errorw("get user orders failed", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required" example:"processing"`
}

// @Summary Update order status (admin)
// @Description Обновляет статус заказа по ID
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param input body order.UpdateStatusRequest true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid order id", "param", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnw("invalid input for update", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	status := enums.OrderStatus(req.Status)
	if !status.IsValid() {
		h.log.Warnw("invalid status value", "status", req.Status)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	err = h.service.AdminUpdateOrderStatus(c.Request.Context(), id, req.Status)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case errors.Is(err, pkgerrors.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
	case err != nil:
		h.log.Errorw("update order failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, gin.H{"message": "status updated"})
	}
}

// @Summary Delete order by ID (admin)
// @Description Удаляет заказ по ID
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid order id", "param", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	err = h.service.DeleteOrder(c.Request.Context(), id)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case err != nil:
		h.log.Errorw("delete order failed", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.Status(http.StatusNoContent)
	}
}

// @Summary Cancel order (user/admin)
// @Description Отмена заказа пользователем
// @Tags orders
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id}/cancel [get]
func (h *Handler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid order id", "param", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(int64)

	err = h.service.CancelOrder(c.Request.Context(), id, userID)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case errors.Is(err, pkgerrors.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, pkgerrors.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "cancel not allowed"})
	case err != nil:
		h.log.Errorw("cancel order failed", "id", id, "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
	}
}

// @Summary Get order stats (admin)
// @Description Получает статистику заказов по статусам за период
// @Tags orders
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to *time.Time
	layout := "2006-01-02"

	if fromStr != "" {
		t, err := time.Parse(layout, fromStr)
		if err != nil {
			h.log.Warnw("invalid from date", "from", fromStr, "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "BAD_REQUEST", "message": "Invalid 'from' date format. Use YYYY-MM-DD."})
			return
		}
		from = &t
	}

	if toStr != "" {
		t, err := time.Parse(layout, toStr)
		if err != nil {
			h.log.Warnw("invalid to date", "to", toStr, "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "BAD_REQUEST", "message": "Invalid 'to' date format. Use YYYY-MM-DD."})
			return
		}
		to = &t
	}

	stats, err := h.service.GetStats(c.Request.Context(), from, to)
	if err != nil {
		h.log.Errorw("failed to get order stats", "from", fromStr, "to", toStr, "error", err)

		switch err {
		case pkgerrors.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"status": "BAD_REQUEST", "message": "Invalid input provided"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": "INTERNAL_ERROR", "message": "Failed to get statistics"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK", "data": stats})
}

// @Summary Export orders (admin)
// @Description Экспорт заказов в JSON по фильтрам
// @Tags orders
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status"
// @Param min_amount query int false "Filter by min amount"
// @Param max_amount query int false "Filter by max amount"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/export [get]
func (h *Handler) Export(c *gin.Context) {
	filter, ok := h.parseExportFilter(c)
	if !ok {
		return
	}

	orders, err := h.service.ExportOrders(c.Request.Context(), filter)
	if err != nil {
		h.log.Errorw("export orders failed", "filter", filter, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// @Summary Export orders as CSV (admin)
// @Description Экспорт заказов в CSV по фильтрам
// @Tags orders
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status"
// @Param min_amount query int false "Filter by min amount"
// @Param max_amount query int false "Filter by max amount"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {string} string "CSV file"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/export/csv [get]
func (h *Handler) ExportCSV(c *gin.Context) {
	filter, ok := h.parseExportFilter(c)
	if !ok {
		return
	}

	orders, err := h.service.ExportOrders(c.Request.Context(), filter)
	if err != nil {
		h.log.Errorw("export csv failed", "filter", filter, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=orders.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	if err := writer.Write([]string{"ID", "UserID", "Status", "DeliveryDate", "PickupPoint", "OrderDate", "TotalAmount", "ReceiptURL", "CreatedAt", "UpdatedAt"}); err != nil {
		h.log.Errorw("write csv header failed", "error", err)
		return
	}

	for _, o := range orders {
		var receipt string
		if o.ReceiptURL != nil {
			receipt = *o.ReceiptURL
		}

		if err := writer.Write([]string{
			strconv.FormatInt(o.ID, 10),
			strconv.FormatInt(o.UserID, 10),
			string(o.Status),
			nullTimeToString(o.DeliveryDate),
			o.PickupPoint,
			o.OrderDate.Format("2006-01-02 15:04:05"),
			strconv.FormatInt(o.TotalAmount, 10),
			receipt,
			o.CreatedAt.Format("2006-01-02 15:04:05"),
			o.UpdatedAt.Format("2006-01-02 15:04:05"),
		}); err != nil {
			h.log.Errorw("write csv row failed", "order_id", o.ID, "error", err)
			return
		}
	}

	if err := writer.Error(); err != nil {
		h.log.Errorw("flush csv writer failed", "error", err)
	}
}

func (h *Handler) parseExportFilter(c *gin.Context) (order.ExportFilter, bool) {
	var filter order.ExportFilter

	if uidStr := c.Query("user_id"); uidStr != "" {
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return filter, false
		}
		filter.UserID = &uid
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := enums.OrderStatus(statusStr)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return filter, false
		}
		filter.Status = &status
	}

	if minStr := c.Query("min_amount"); minStr != "" {
		min, err := strconv.ParseInt(minStr, 10, 64)
		if err != nil || min < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_amount"})
			return filter, false
		}
		filter.MinAmount = &min
	}

	if maxStr := c.Query("max_amount"); maxStr != "" {
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil || max < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max_amount"})
			return filter, false
		}
		filter.MaxAmount = &max
	}

	if filter.MinAmount != nil && filter.MaxAmount != nil && *filter.MinAmount > *filter.MaxAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount must be <= max_amount"})
		return filter, false
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return filter, false
		}
		filter.Limit = limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return filter, false
		}
		filter.Offset = offset
	}

	return filter, true
}

func nullTimeToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
