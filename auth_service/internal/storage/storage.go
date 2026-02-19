package storage

import (

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
	"github.com/jmoiron/sqlx"
)

type UserOperations interface {
}

type Storage struct{
	UserOperations
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{
		UserOperations: postgres.NewUserOperationsPostgres(db),
	}
}