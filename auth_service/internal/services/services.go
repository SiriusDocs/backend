package services

import (
	"log/slog"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services/users"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
)

type UserOperations interface {
	CreateUser(username string, email string, password string) (int64, error)
}

type Service struct {
	UserOperations
}

func NewService(logger *slog.Logger, store *storage.Storage) *Service {
	return &Service{
		UserOperations: users.NewUserOperations(logger, store.UserOperations),
	}
}
