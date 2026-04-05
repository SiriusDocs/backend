package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SiriusDocs/backend/api_gateway/internal/app"
	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/pkg/logger"
)

// @title           Sirius docs API
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main(){
	cfg := config.MustLoad()
	logger := logger.SetupLogger(cfg.Env)

	application, err := app.New(logger, cfg)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := application.HTTPServer.Run(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("error occured while running http server: %s", err.Error()))
		}
	}()


	logger.Info("App Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("App Shutting Down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.HTTPServer.Stop(ctx); err != nil {
		logger.Error(fmt.Sprintf("error occured on server shutting down: %s", err.Error()))
	}
	logger.Info("Gracefully stopped")
}
