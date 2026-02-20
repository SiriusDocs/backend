package storage

import (
	"context"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	"github.com/jmoiron/sqlx"
)

type UserOperations interface {
	CreateUser(ctx context.Context, username string, email string, password string) (int64, error)
	GetUser(ctx context.Context, email string, password string) (domain.User, error)
	SetSession(ctx context.Context, userId int64, session tokenmanager.Session) error
	IsTokenValid(ctx context.Context, refreshToken string) (int64, error)
}

type Storage struct {
	UserOperations
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{
		UserOperations: postgres.NewUserOperationsPostgres(db),
	}
}
