package auth

import (
    "log/slog"
    "github.com/gin-gonic/gin"
    "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/api_gateway/internal/services"
)

type Handler struct {
    log     *slog.Logger
    service services.AuthService
}

// Конструктор хэндлера получает сервис
func New(log *slog.Logger, service services.AuthService) *Handler {
    return &Handler{
        log:     log,
        service: service,
    }
}

// Метод для регистрации роутов ИМЕННО этого домена
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
    authGroup := rg.Group("/auth")
    {
        authGroup.POST("/sign-up", h.signUp)
        authGroup.POST("/sign-in", h.signIn)
    }
}

