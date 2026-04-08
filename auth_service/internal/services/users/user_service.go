package users

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log/slog"
	"os"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 87600 * time.Minute
)

type UsersService struct{
	log *slog.Logger
	store storage.UserOperations
	manager *tokenmanager.Manager
}

func NewUserOperations(logger *slog.Logger, store storage.UserOperations, manager *tokenmanager.Manager) *UsersService {
	return &UsersService{
		log: logger,
		store: store,
		manager: manager,
	}
}

func (u *UsersService) CreateUser(ctx context.Context, username string, email string, password string) (int64, error){
	id, err := u.store.CreateUser(ctx, username,email,generatePasswordHash(password))
	return id, err
}

func (u *UsersService) GenerateTokens(ctx context.Context, email string, password string) (int64, domain.Tokens, error) {
	user, err := u.store.GetUser(ctx, email, generatePasswordHash(password))
	if err != nil {
		fmt.Println(err)
		return 0,domain.Tokens{},err
	}
	return u.createSession(ctx, user)
}

func (u *UsersService) RefreshToken(ctx context.Context, refreshToken string) (domain.Tokens, error) {
	userId, err := u.store.IsTokenValid(ctx, refreshToken)
	if err != nil {
		return domain.Tokens{}, err
	}

	// Обязательно достаем пользователя, чтобы узнать его актуальную роль (вдруг админ ее обновил)
	user, err := u.store.GetUserById(ctx, userId)
	if err != nil {
		return domain.Tokens{}, err
	}

	_, tokens, err := u.createSession(ctx, user)
	return tokens, err
}

func (u *UsersService) GetProfile(ctx context.Context, userId int64) (domain.User, error) {
	return u.store.GetUserById(ctx, userId)
}

func (u *UsersService) GetPendingUsers(ctx context.Context, limit, offset int32) ([]domain.User, int32, error) {
	return u.store.GetPendingUsers(ctx, limit, offset)
}

func (u *UsersService) AssignRole(ctx context.Context, targetUserId int64, newRole string) error {
	return u.store.UpdateUserRole(ctx, targetUserId, newRole)
}

//--------------------

func (u *UsersService) createSession(ctx context.Context, user domain.User) (int64, domain.Tokens, error) {
	var (
		res domain.Tokens
		err error
	)
	res.AccessToken, err = u.manager.NewJWT(user, accessTokenTTL)
	if err != nil {
		return 0, domain.Tokens{}, err
	}
	res.RefreshToken, err = u.manager.NewRefreshToken()
	if err != nil {
		return 0, domain.Tokens{}, err
	}

	session := tokenmanager.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(refreshTokenTTL),
	}
	err = u.store.SetSession(ctx, user.Id, session)
	return user.Id, res, err
}


func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(os.Getenv("SALT"))))

}
