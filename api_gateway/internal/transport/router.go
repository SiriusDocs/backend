package transport

import (
	"time"

	_ "github.com/SiriusDocs/backend/api_gateway/docs"
	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 1
	router.RedirectTrailingSlash = false

	config := cors.DefaultConfig()

	if len(cfg.HTTPServer.AllowedOrigins) > 0 && cfg.HTTPServer.AllowedOrigins[0] != "" {
		config.AllowOrigins = cfg.HTTPServer.AllowedOrigins
	} else {
		config.AllowOrigins = []string{"http://localhost:5174"}
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

