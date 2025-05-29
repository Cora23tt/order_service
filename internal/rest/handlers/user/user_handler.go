package user

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Cora23tt/order_service/internal/usecase/user"
	pkgerrors "github.com/Cora23tt/order_service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *user.Service
	log     *zap.SugaredLogger
}

func NewHandler(service *user.Service, log *zap.SugaredLogger) *Handler {
	return &Handler{service: service, log: log}
}

// GetProfile godoc
// @Summary Получение профиля текущего пользователя
// @Description Возвращает информацию о текущем пользователе по JWT-токену
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} user.Profile "Профиль пользователя"
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 404 {object} map[string]string "user not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /api/v1/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(int64)

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		h.log.Errorw("get profile failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

type UpdateProfileRequest struct {
	PINFL     *string `json:"pinfl,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateProfile godoc
// @Summary Обновление профиля пользователя
// @Description Обновляет PINFL и аватар текущего пользователя
// @Tags User
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param pinfl formData string false "ПИНФЛ пользователя"
// @Param avatar formData file false "Аватар (изображение)"
// @Success 200 {object} map[string]string "profile updated"
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 404 {object} map[string]string "user not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /api/v1/me [patch]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(int64)

	var pinfl string
	if val := c.PostForm("pinfl"); val != "" {
		pinfl = val
	}

	file, err := c.FormFile("avatar")
	var avatarPath *string
	if err == nil {
		ext := filepath.Ext(file.Filename)
		baseDir := filepath.Join(".", "web", "avatars", fmt.Sprintf("%d", userID))

		if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
			h.log.Errorw("failed to create avatar dir", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create folder"})
			return
		}

		filename := "profile" + ext
		fullPath := filepath.Join(baseDir, filename)

		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			h.log.Errorw("failed to save avatar", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
			return
		}
		url := fmt.Sprintf("/profile/%d/photo", userID)
		avatarPath = &url
	}

	var pinflPtr *string
	if pinfl != "" {
		pinflPtr = &pinfl
	}

	if err := h.service.UpdateProfile(c.Request.Context(), userID, pinflPtr, avatarPath); err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		h.log.Errorw("update profile failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// ListUsers godoc
// @Summary Получение списка всех пользователей (admin)
// @Description Админский доступ. Возвращает список всех зарегистрированных пользователей
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {array} user.Profile
// @Failure 403 {object} map[string]string "forbidden"
// @Failure 500 {object} map[string]string "internal error"
// @Router /api/v1/admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	users, err := h.service.ListAllUsers(c.Request.Context())
	if err != nil {
		h.log.Errorw("list users failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetProfilePhoto godoc
// @Summary Получение аватара пользователя
// @Description Возвращает файл изображения профиля по ID пользователя
// @Tags User
// @Produce image/jpeg
// @Produce image/png
// @Produce image/webp
// @Param id path int true "ID пользователя"
// @Success 200 {file} file "Изображение аватара"
// @Failure 404 {object} map[string]string "photo not found"
// @Router /profile/{id}/photo [get]
func (h *Handler) GetProfilePhoto(c *gin.Context) {
	id := c.Param("id")
	exts := []string{".jpg", ".jpeg", ".png", ".webp"}
	for _, ext := range exts {
		path := fmt.Sprintf("web/avatars/%s/profile%s", id, ext)
		if _, err := os.Stat(path); err == nil {
			c.File(path)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "photo not found"})
}
