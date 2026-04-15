package files

import (
	"log/slog"

	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     *slog.Logger
	service services.FileService
	client  *config.GRPCClient
}

func New(log *slog.Logger, service services.FileService, client *config.GRPCClient) *Handler {
	return &Handler{
		log:     log,
		service: service,
		client:  client,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	filesGroup := rg.Group("/files")
	{
		// User routes
		filesGroup.POST("/upload", h.uploadFile)
		filesGroup.GET("/download/:file_id", h.downloadFile)
		filesGroup.GET("/status/:upload_id", h.getUploadStatus)
		filesGroup.GET("/meta/:file_id", h.getFileMeta)

		// Admin routes
		adminGroup := filesGroup.Group("/admin")
		{
			adminGroup.POST("/buckets", h.createBucket)
			adminGroup.DELETE("/:file_id", h.deleteFile)
			adminGroup.GET("/list/:bucket", h.listFiles)
		}
	}
}