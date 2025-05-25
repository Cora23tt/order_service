package auth

import (
	"errors"
	"net/http"

	"github.com/Cora23tt/order_service/internal/usecase/auth"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	service *auth.Service
	log     *zap.SugaredLogger
}

func NewHandler(service *auth.Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) SignUp(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		http.Error(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(c.Writer, "Error hashing password", http.StatusInternalServerError)
		return
	}

	userID, err := h.service.CreateUser(c.Request.Context(), creds.Username, string(hashedPassword))
	if err != nil {
		switch {
		case errors.Is(err, pkgerrors.ErrAlreadyExists):
			http.Error(c.Writer, "Username already taken", http.StatusConflict)
		default:
			http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	token, err := h.service.CreateSessionToken(c.Request.Context(), userID)
	if err != nil {
		http.Error(c.Writer, "Error initialising session token", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user created", "token": token})
}

func (h *Handler) SignIn(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		http.Error(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByUsername(c.Request.Context(), creds.Username)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			http.Error(c.Writer, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(creds.Password)) != nil {
		http.Error(c.Writer, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.service.CreateSessionToken(c.Request.Context(), user.ID)
	if err != nil {
		http.Error(c.Writer, "Error initialising session token", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) Me(c *gin.Context) {

	userID, ok := c.Get("userID")
	if !ok {
		http.Error(c.Writer, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(c.Writer, "Invalid user ID type", http.StatusInternalServerError)
		return
	}

	user, err := h.service.GetUserByID(c, userIDStr)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			http.Error(c.Writer, "User not found", http.StatusNotFound)
		} else {
			http.Error(c.Writer, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}
