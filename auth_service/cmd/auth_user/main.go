package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/app"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	logger := logger.SetupLogger(cfg.Env)
	
	application := app.New(logger, cfg)

	go func(){
		application.GRPCServer.MustRun()
	}()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM
    <-stop

    // initiate graceful shutdown
    application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
    //application.Storage.Stop()
    logger.Info("Gracefully stopped")    
}