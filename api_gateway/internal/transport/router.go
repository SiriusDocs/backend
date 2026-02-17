package transport

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
    _ "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/api_gateway/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

)

// Router больше не содержит логику бизнес-хэндлеров
func NewRouter() *gin.Engine {
    router := gin.Default()
    router.MaxMultipartMemory = 1
    router.RedirectTrailingSlash = false

    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:5174"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))

    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}