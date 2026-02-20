package storage

import (

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
	"github.com/jmoiron/sqlx"
)

type UserOperations interface {
	CreateUser(username string, email string, password string) (int64, error)
}

type Storage struct{
	UserOperations
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{
		UserOperations: postgres.NewUserOperationsPostgres(db),
	}
}