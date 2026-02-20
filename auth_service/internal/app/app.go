package app

import (
	"log/slog"

	grpcapp "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/app/grpc"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
		_ "github.com/lib/pq"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(logger *slog.Logger, cfg *config.Config) *App {

	db, err := postgres.NewPostgresDB(cfg.Db.Host, cfg.Db.Port, cfg.Db.Username, cfg.Db.DBName, cfg.Db.Password, cfg.Db.SSLMode)
	if err != nil {
		panic(err)
	}
	storage := storage.NewPostgresStorage(db)
	service := services.NewService(logger, storage)
	grpcApp := grpcapp.New(logger, service, cfg.GRPC.Port)

	return &App{GRPCServer: grpcApp}
}