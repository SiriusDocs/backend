package postgres

import (
	// "context"
	// "database/sql"
	// "errors"
	// "fmt"
	// "time"
	//
	// "github.com/SiriusDocs/backend/fill_service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type FillStoragePostgres struct {
	db *sqlx.DB
}

func NewFillStoragePostgres(db *sqlx.DB) *FillStoragePostgres {
	return &FillStoragePostgres{
		db: db,
	}
}
