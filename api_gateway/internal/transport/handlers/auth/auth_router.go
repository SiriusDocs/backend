package auth

import (
	"log/slog"

	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     *slog.Logger
	service services.AuthService
	client  *config.GRPCClient
}

func New(log *slog.Logger, service services.AuthService, client  *config.GRPCClient) *Handler {
	return &Handler{
		log:     log,
		service: service,
        client: client,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/sign-up", h.signUp)
		authGroup.POST("/sign-in", h.signIn)
		authGroup.POST("/refresh", h.refreshToken)
	}
}
