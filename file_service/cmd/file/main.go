package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/app"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/config"
	logs "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	logger := logs.SetupLogger(cfg.Env)
	logger = logger.With(logs.BaseAttrs(cfg.Env)...)
	application := app.New(logger, cfg)
	go func(){
		application.GRPCServer.MustRun()
	}()
	stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	// Waiting for SIGINT (pkill -2) or SIGTERM
    <-stop
    // initiate graceful shutdown
    application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
    logger.Info("Gracefully stopped") 
}
