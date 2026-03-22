package templates

import (
	"log/slog"

	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     *slog.Logger
	service services.TempService
	client  *config.GRPCClient
}

func New(log *slog.Logger, service services.TempService, client  *config.GRPCClient) *Handler {
	return &Handler{
		log:     log,
		service: service,
        client: client,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	tempGroup := rg.Group("/temp")
	{
		fileGroup := tempGroup.Group("/files")
		{
			fileGroup.POST("/upload", h.uploadFile)
			fileGroup.POST("/status/:task_id", h.checkStatus)
		}
	}
}