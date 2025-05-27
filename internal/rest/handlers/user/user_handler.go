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

type UpdateProfileRequest struct {
	PINFL     *string `json:"pinfl,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

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
