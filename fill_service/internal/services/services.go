package fill

import (
	// "context"
	"log/slog"

	// "github.com/SiriusDocs/backend/fill_service/internal/config"
	fill_service "github.com/SiriusDocs/backend/fill_service/internal/services/fill"
	"github.com/SiriusDocs/backend/fill_service/internal/storage"
)

// Бизнес-логика самого заполнения документов
type FillService interface {
}

type Service struct {
	FillService
}

func NewService(log *slog.Logger, store *storage.Storage) *Service {
	return &Service{
		FillService: fill_service.NewFillService(log, store.FillStorage),
	}
}
