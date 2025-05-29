package product

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	productService "github.com/Cora23tt/order_service/internal/usecase/product"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
)

type Handler struct {
	service *productService.Service
	log     *zap.SugaredLogger
}

func NewHandler(service *productService.Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

type Product struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Price       int64  `json:"price" binding:"required,gte=0"`
	Quantity    int64  `json:"quantity" binding:"required,gte=0"`
}

// @Summary Get all products
// @Description Возвращает список всех доступных продуктов
// @Tags products
// @Success 200 {object} map[string]interface{} "products: []Product"
// @Failure 500 {object} map[string]string
// @Router /api/v1/products [get]
func (h *Handler) GetProducts(c *gin.Context) {
	products, err := h.service.GetProducts(c.Request.Context())
	if err != nil {
		h.log.Errorw("failed to get products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	h.log.Infow("product list retrieved", "count", len(products))
	c.JSON(http.StatusOK, gin.H{"products": products})
}

// @Summary Get product by ID
// @Description Возвращает продукт по его ID
// @Tags products
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{} "product: Product"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid product id for get", "raw", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		h.log.Warnw("product not found", "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
	case err != nil:
		h.log.Errorw("get product error", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		h.log.Infow("product retrieved", "id", id)
		c.JSON(http.StatusOK, gin.H{"product": product})
	}
}

// @Summary Add new product (admin)
// @Description Добавляет новый продукт в каталог
// @Tags products
// @Security BearerAuth
// @Param product body product.Product true "Product object"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products [post]
func (h *Handler) AddProduct(c *gin.Context) {
	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		h.log.Warnw("invalid JSON for add product", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.AddProduct(c.Request.Context(), p.Price, p.Quantity, p.Name, p.Description, p.ImageURL)
	switch {
	case errors.Is(err, pkgerrors.ErrInvalidInput):
		h.log.Warnw("invalid input for add product", "name", p.Name, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, pkgerrors.ErrAlreadyExists):
		h.log.Warnw("duplicate product", "name", p.Name)
		c.JSON(http.StatusConflict, gin.H{"error": "product already exists"})
	case err != nil:
		h.log.Errorw("failed to add product", "name", p.Name, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		h.log.Infow("product added", "name", p.Name)
		c.JSON(http.StatusCreated, gin.H{"message": "Product added"})
	}
}

// @Summary Update product by ID (admin)
// @Description Обновляет информацию о продукте по ID
// @Tags products
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product body product.Product true "Updated product"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid product id", "raw", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		h.log.Warnw("invalid JSON for update product", "id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateProduct(c.Request.Context(), id, p.Quantity, p.Price, p.Name, p.Description, p.ImageURL)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		h.log.Warnw("product not found for update", "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
	case errors.Is(err, pkgerrors.ErrInvalidInput):
		h.log.Warnw("invalid input for update product", "id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, pkgerrors.ErrAlreadyExists):
		h.log.Warnw("product update conflict", "id", id)
		c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
	case err != nil:
		h.log.Errorw("failed to update product", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		h.log.Infow("product updated", "id", id)
		c.JSON(http.StatusOK, gin.H{"message": "Product updated"})
	}
}

// @Summary Delete product by ID (admin)
// @Description Удаляет продукт по ID
// @Tags products
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Warnw("invalid product id for delete", "raw", c.Param("id"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.service.DeleteProduct(c.Request.Context(), id)
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		h.log.Warnw("product not found for delete", "id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
	case err != nil:
		h.log.Errorw("failed to delete product", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		h.log.Infow("product deleted", "id", id)
		c.Status(http.StatusNoContent)
	}
}
