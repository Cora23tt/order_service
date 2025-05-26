package middleware

import (
	"go.uber.org/zap"
)

type Middleware struct {
	logger    *zap.SugaredLogger
	validator AuthValidator
}

func New(logger *zap.SugaredLogger, validator AuthValidator) *Middleware {
	return &Middleware{
		logger:    logger,
		validator: validator,
	}
}
