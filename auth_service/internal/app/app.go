package app

import (
	"fmt"
	"log/slog"

	grpcapp "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/app/grpc"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	fileclient "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/clients/file"
	_ "github.com/lib/pq"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(logger *slog.Logger, cfg *config.Config) (*App, error) {
	const op = "app.New"

	db, err := postgres.NewPostgresDB(cfg.Db.Host, cfg.Db.Port, cfg.Db.Username, cfg.Db.DBName, cfg.Db.Password, cfg.Db.SSLMode)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to initialize postgres storage: %w", op, err)
	}

	manager, err := tokenmanager.NewManager(cfg.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to initialize token manager: %w", op, err)
	}

	// инициализация gRPC подключения к FileService
	fileConn, err := grpcapp.NewClient(cfg.FileService.Address, logger)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect file_service: %w", op, err)
	}
	fileClient := fileclient.New(fileConn)

	storage := storage.NewPostgresStorage(db)
	service := services.NewService(logger, storage, manager, fileClient)
	grpcApp := grpcapp.New(logger, service, cfg.GRPC.Port)

	return &App{GRPCServer: grpcApp}, nil
}
