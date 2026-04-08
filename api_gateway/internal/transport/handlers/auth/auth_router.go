package auth

import (
	"log/slog"

	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/SiriusDocs/backend/api_gateway/internal/transport/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     *slog.Logger
	service services.AuthService
	client  *config.GRPCClient
	jwtSecret string
}

func New(log *slog.Logger, service services.AuthService, client *config.GRPCClient, jwtSecret string) *Handler {
	return &Handler{
		log:       log,
		service:   service,
		client:    client,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/sign-up", h.signUp)
		authGroup.POST("/sign-in", h.signIn)
		authGroup.POST("/refresh", h.refreshToken)

		authGroup.POST("/")
	}
	profileGroup := rg.Group("/profile", middleware.UserIdentity(h.jwtSecret))
	{
		profileGroup.GET("/me", h.getProfile) 
	}
	
	adminGroup := rg.Group("/admin", middleware.UserIdentity(h.jwtSecret), middleware.RequireRole("ADMIN"))
	{
		adminGroup.GET("/pending-users", h.listPendingUsers)
		adminGroup.POST("/assign-role", h.assignRole)
	}
}

