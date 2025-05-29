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
	PhoneNumber string `json:"phone_number" example:"+998901234567"`
	Password    string `json:"password" example:"strong_password_123"`
}

// SignUp godoc
// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя по номеру телефона и паролю
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body Credentials true "Данные пользователя"
// @Success 200 {object} map[string]int64 "ID нового пользователя"
// @Failure 400 {string} string "Invalid request body"
// @Failure 409 {string} string "A user with this phone number already exists"
// @Failure 500 {string} string "Internal server error"
// @Router /api/v1/auth/signup [post]
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

// SignIn godoc
// @Summary Вход пользователя
// @Description Аутентифицирует пользователя и выдает JWT токен
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body Credentials true "Телефон и пароль"
// @Success 200 {object} map[string]string "JWT токен"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Error initialising session token"
// @Router /api/v1/auth/signin [post]
func (h *Handler) SignIn(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		http.Error(c.Writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.service.GenerateJWT(c.Request.Context(), creds.PhoneNumber, creds.Password)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) || errors.Is(err, pkgerrors.ErrInvalidCredentials) {
			http.Error(c.Writer, "Invalid phone number or password", http.StatusUnauthorized)
			return
		}
		http.Error(c.Writer, "Error initialising session token", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
