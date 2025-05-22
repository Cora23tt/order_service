package middleware

import (
	"github.com/Cora23tt/order_service/internal/usecase/auth"
	"go.uber.org/zap"
)

type Middleware struct {
	logger      *zap.SugaredLogger
	authService *auth.Service
}

func New(logger *zap.SugaredLogger, authService *auth.Service) *Middleware {
	return &Middleware{
		logger:      logger,
		authService: authService,
	}
}
