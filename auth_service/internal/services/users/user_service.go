package users

import (
	"log/slog"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
)

type UsersService struct{
	log *slog.Logger
	store storage.UserOperations
}

func NewUserOperations(logger *slog.Logger, store storage.UserOperations) *UsersService {
	return &UsersService{
		log: logger,
		store: store,
	}
}

func (u *UsersService) CreateUser() {}