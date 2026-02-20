package users

import (
	"crypto/sha1"
	"fmt"
	"log/slog"
	"os"

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

func (u *UsersService) CreateUser(username string, email string, password string) (int64, error){
	id, err := u.store.CreateUser(username,email,generatePasswordHash(password))
	return id, err
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(os.Getenv("SALT"))))

}
