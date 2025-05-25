package auth

import (
	"errors"
	"net/http"

	"github.com/Cora23tt/order_service/internal/usecase/auth"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *auth.Service
	log     *zap.SugaredLogger
}

func NewHandler(service *auth.Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

type Credentials struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

func (h *Handler) SignUp(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		http.Error(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.service.CreateUser(c.Request.Context(), creds.PhoneNumber, creds.Password)
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrAlreadyExists):
			http.Error(c.Writer, "A user with this phone number already exists", http.StatusConflict)
		default:
			http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID})
}

func (h *Handler) SignIn(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		http.Error(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.service.GenerateJWT(c.Request.Context(), creds.PhoneNumber, creds.Password)
	if err != nil {
		http.Error(c.Writer, "Error initialising session token", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
