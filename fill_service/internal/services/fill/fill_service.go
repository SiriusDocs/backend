package fill

import (
	// "context"
	// "encoding/json"
	// "fmt"
	"log/slog"
	// "strings"

	// "github.com/SiriusDocs/backend/fill_service/internal/domain"
	"github.com/SiriusDocs/backend/fill_service/internal/storage"
)

type FillService struct {
	log   *slog.Logger
	store storage.FillStorage
}

func NewFillService(log *slog.Logger, store storage.FillStorage) *FillService {
	return &FillService{
		log:   log,
		store: store,
	}
}
