package services

import (
	"context"
	"log/slog"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services/users"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
)

type UserOperations interface {
	CreateUser(ctx context.Context, username string, email string, password string) (int64, error)
	GenerateTokens(ctx context.Context, email string, password string) (int64, domain.Tokens, error)
}

type Service struct {
	UserOperations
}

func NewService(logger *slog.Logger, store *storage.Storage, manager *tokenmanager.Manager) *Service {
	return &Service{
		UserOperations: users.NewUserOperations(logger, store.UserOperations, manager),
	}
}
