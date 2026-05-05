package storage

import (
	// "context"
	//
	// "github.com/SiriusDocs/backend/fill_service/internal/domain"
	"github.com/SiriusDocs/backend/fill_service/internal/storage/postgres"
	"github.com/jmoiron/sqlx"
)

type FillStorage interface {
}

type Storage struct {
	FillStorage
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{
		FillStorage: postgres.NewFillStoragePostgres(db),
	}
}
