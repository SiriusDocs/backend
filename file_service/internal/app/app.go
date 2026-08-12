package app

import (
	"context"
	"fmt"
	"log/slog"

	grpcapp "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/app/grpc"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/services"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/storage"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*App, error) {
	const op = "app.New"

	st, err := storage.NewS3Storage(ctx, cfg.S3)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to initialize s3 storage: %w", op, err)
	}
	service := services.NewService(logger,st)
	grpcapp := grpcapp.New(logger, service, cfg.GRPC.Port, cfg.GRPC.Timeout)

	return &App{ GRPCServer: grpcapp}, nil
}