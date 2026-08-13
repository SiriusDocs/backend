package users

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	logs "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/logger"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	"github.com/google/uuid"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 87600 * time.Minute
	ttl = int64(15) // TTL 15 минут
)

type FileServiceClient interface {
	GenerateUploadURL(ctx context.Context, key, contentType string, ttlMinutes int64) (string, error)
	GenerateDownloadURL(ctx context.Context, key string, ttlMinutes int64) (string, error)
}

type UsersService struct{
	log *slog.Logger
	store storage.UserOperations
	manager *tokenmanager.Manager
	fileService FileServiceClient
}

func NewUserOperations(logger *slog.Logger, store storage.UserOperations, manager *tokenmanager.Manager, fileService FileServiceClient) *UsersService {
	return &UsersService{
		log: logger,
		store: store,
		manager: manager,
		fileService: fileService,
	}
}

func (u *UsersService) CreateUser(ctx context.Context, username string, email string, password string) (int64, error) {
	const op = "services.users.CreateUser"
	start := time.Now()

	id, err := u.store.CreateUser(ctx, username, email, generatePasswordHash(password))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		// если пользователь уже существует — это предупреждение (WARN)
		if errors.Is(err, domain.ErrUserExists) {
			u.log.WarnContext(ctx, "user registration failed: already exists",
				slog.String("username", username),
				slog.String("email", email),
				logs.ErrAttrs(wrappedErr),
				logs.ResultAttrs("error", start),
			)
			return 0, wrappedErr
		}

		u.log.ErrorContext(ctx, "failed to create user in database",
			slog.String("username", username),
			slog.String("email", email),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)
		return 0, wrappedErr
	}

	u.log.InfoContext(ctx, "user created successfully",
		slog.Int64("user_id", id),
		slog.String("username", username),
		slog.String("email", email),
		logs.ResultAttrs("ok", start),
	)

	return id, nil
}

func (u *UsersService) GenerateTokens(ctx context.Context, email string, password string) (int64, domain.Tokens, error) {
	const op = "services.users.GenerateTokens"
	start := time.Now()

	// ищем пользователя по логину и хэшу пароля
	user, err := u.store.GetUser(ctx, email, generatePasswordHash(password))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		// неверные учетные данные — это бизнес-предупреждение (WARN)
		if errors.Is(err, domain.ErrInvalidCreds) || errors.Is(err, domain.ErrUserNotFound) {
			u.log.WarnContext(ctx, "invalid login attempt",
				slog.String("email", email),
				logs.ErrAttrs(wrappedErr),
				logs.ResultAttrs("error", start),
			)
			return 0, domain.Tokens{}, wrappedErr
		}

		u.log.ErrorContext(ctx, "failed to get user during login",
			slog.String("email", email),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)
		return 0, domain.Tokens{}, wrappedErr
	}

	// cоздаем сессию и токены
	userID, tokens, err := u.createSession(ctx, user)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		u.log.ErrorContext(ctx, "failed to create user session during login",
			slog.Int64("user_id", user.Id),
			slog.String("email", email),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)
		return 0, domain.Tokens{}, wrappedErr
	}

	u.log.InfoContext(ctx, "user logged in successfully",
		slog.Int64("user_id", userID),
		slog.String("email", email),
		logs.ResultAttrs("ok", start),
	)

	return userID, tokens, nil
}

func (u *UsersService) RefreshToken(ctx context.Context, refreshToken string) (domain.Tokens, error) {
	const op = "services.users.RefreshToken"
	start := time.Now()

	// проверяем валидность refresh-токена в базе
	userID, err := u.store.IsTokenValid(ctx, refreshToken)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		u.log.WarnContext(ctx, "failed to refresh tokens: invalid or expired refresh token",
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)
		return domain.Tokens{}, wrappedErr
	}

	// перевыпускаем сессию для пользователя
	user := domain.User{Id: userID}
	_, tokens, err := u.createSession(ctx, user)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)

		u.log.ErrorContext(ctx, "failed to create session during token refresh",
			slog.Int64("user_id", userID),
			logs.ErrAttrs(wrappedErr),
			logs.ResultAttrs("error", start),
		)
		return domain.Tokens{}, wrappedErr
	}

	u.log.InfoContext(ctx, "tokens refreshed successfully",
		slog.Int64("user_id", userID),
		logs.ResultAttrs("ok", start),
	)

	return tokens, nil
}


// GetProfile возвращает данные профиля пользователя с рабочей ссылкой на аватар
func (u *UsersService) GetProfile(ctx context.Context, userID int64) (domain.User, string, error) {
	const op = "services.users.GetProfile"
	start := time.Now()

	user, err := u.store.GetUserByID(ctx, userID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		u.log.ErrorContext(ctx, "failed to get user profile", slog.Int64("user_id", userID), logs.ErrAttrs(wrappedErr))
		return domain.User{}, "", wrappedErr
	}

	var avatarURL string
	if user.AvatarKey != "" && u.fileService != nil {
		avatarURL, err = u.fileService.GenerateDownloadURL(ctx, user.AvatarKey, ttl) 
		if err != nil {
			// Логируем ошибку S3, но профиль всё равно отдаем без аватарки
			u.log.WarnContext(ctx, "failed to generate avatar download url", slog.String("key", user.AvatarKey), logs.ErrAttrs(err))
		}
	}

	u.log.InfoContext(ctx, "user profile retrieved", slog.Int64("user_id", userID), logs.ResultAttrs("ok", start))
	return user, avatarURL, nil
}

// GetAvatarURL возвращает прямую presigned-ссылку на аватар пользователя
func (u *UsersService) GetAvatarURL(ctx context.Context, userID int64) (string, error) {
	const op = "services.users.GetAvatarURL"
	start := time.Now()

	user, err := u.store.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if user.AvatarKey == "" {
		return "", fmt.Errorf("%s: %w", op, domain.ErrAvatarNotFound)
	}

	avatarURL, err := u.fileService.GenerateDownloadURL(ctx, user.AvatarKey, ttl)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		u.log.ErrorContext(ctx, "failed to generate avatar url", slog.Int64("user_id", userID), logs.ErrAttrs(wrappedErr))
		return "", wrappedErr
	}

	u.log.InfoContext(ctx, "avatar url generated", slog.Int64("user_id", userID), logs.ResultAttrs("ok", start))
	return avatarURL, nil
}

// GenerateAvatarUploadURL формирует ключ аватара, обновляет его в DB и отдает Presigned PUT ссылку
func (u *UsersService) GenerateAvatarUploadURL(ctx context.Context, userID int64, contentType string) (string, string, error) {
	const op = "services.users.GenerateAvatarUploadURL"
	start := time.Now()

	// формируем уникальный S3 Key: avatars/user_123_<uuid>.png
	avatarKey := fmt.Sprintf("avatars/user_%d_%s", userID, uuid.New().String())

	// генерируем presigned ссылку на загрузку в S3
	uploadURL, err := u.fileService.GenerateUploadURL(ctx, avatarKey, contentType, 15)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		u.log.ErrorContext(ctx, "failed to generate avatar upload url", slog.Int64("user_id", userID), logs.ErrAttrs(wrappedErr))
		return "", "", wrappedErr
	}

	// сохраняем новый avatarKey в БД пользователя
	if err := u.store.UpdateAvatarKey(ctx, userID, avatarKey); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		u.log.ErrorContext(ctx, "failed to update user avatar key in db", slog.Int64("user_id", userID), logs.ErrAttrs(wrappedErr))
		return "", "", wrappedErr
	}

	u.log.InfoContext(ctx, "avatar upload url generated", slog.Int64("user_id", userID), slog.String("key", avatarKey), logs.ResultAttrs("ok", start))
	return uploadURL, avatarKey, nil
}

//--------------------

func (u *UsersService) createSession(ctx context.Context, user domain.User) (int64, domain.Tokens, error) {
	const op = "services.users.createSession"
	start := time.Now()

	var (
		res domain.Tokens
		err error
	)

	// Генерация Access JWT
	res.AccessToken, err = u.manager.NewJWT(user, accessTokenTTL)
	if err != nil {
		return 0, domain.Tokens{}, fmt.Errorf("%s: generate jwt: %w", op, err)
	}

	// Генерация Refresh токена
	res.RefreshToken, err = u.manager.NewRefreshToken()
	if err != nil {
		return 0, domain.Tokens{}, fmt.Errorf("%s: generate refresh token: %w", op, err)
	}

	// Сохранение сессии в Postgres
	session := tokenmanager.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(refreshTokenTTL),
	}

	err = u.store.SetSession(ctx, user.Id, session)
	if err != nil {
		return 0, domain.Tokens{}, fmt.Errorf("%s: save session: %w", op, err)
	}

	u.log.DebugContext(ctx, "session created and saved to storage",
		slog.Int64("user_id", user.Id),
		logs.ResultAttrs("ok", start),
	)

	return user.Id, res, nil
}


func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(os.Getenv("SALT"))))

}
