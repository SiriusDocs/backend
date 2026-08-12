package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/app"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/config"
	logs "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	logger := logs.SetupLogger(cfg.Env)
	logger = logger.With(logs.BaseAttrs(cfg.Env)...)

	// Контекст для инициализации приложения с таймаутом 10 сек
    initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

	application, err := app.New(initCtx, logger, cfg)
	if err != nil {
		logger.Error("failed to initialize application", logs.ErrAttrs(err))
		os.Exit(1)
	}
	go func(){
		application.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	// Waiting for SIGINT (pkill -2) or SIGTERM
    <-stop
    // initiate graceful shutdown
    application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
}
