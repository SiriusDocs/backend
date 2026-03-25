package app

import (
	"log/slog"

	grpcapp "github.com/SiriusDocs/backend/template_service/internal/app/grpc"
	"github.com/SiriusDocs/backend/template_service/internal/config"
	"github.com/SiriusDocs/backend/template_service/internal/services"
	"github.com/SiriusDocs/backend/template_service/internal/storage"
	"github.com/SiriusDocs/backend/template_service/internal/storage/postgres"
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
	service := templates.NewService(logger, storage)
	grpcApp := grpcapp.New(logger, service, cfg.GRPC.Port)

	return &App{GRPCServer: grpcApp}
}