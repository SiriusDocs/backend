package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserOperationsPostgres struct {
	db *sqlx.DB
}

func NewUserOperationsPostgres(db *sqlx.DB) *UserOperationsPostgres {
	return &UserOperationsPostgres{
		db: db,
	}
}

func (u *UserOperationsPostgres) CreateUser(ctx context.Context, username string, email string, password string) (int64, error) {
	var id int64
	query := fmt.Sprintf("INSERT INTO %s (username, email, hashed_password) values ($1, $2, $3) RETURNING id", usersTable)

	if err := u.db.GetContext(ctx, &id, query, username, email, password); err != nil {
		// Unique Violation
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // 23505 — код unique_violation
				return 0, domain.ErrUserExists
			}
		}
		return 0, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (u *UserOperationsPostgres) GetUser(ctx context.Context, email string, password string) (domain.User, error) {
	var user domain.User
	query := fmt.Sprintf("SELECT id,username,email,creation_timestamp FROM %s WHERE email=$1 AND hashed_password=$2", usersTable)
	if err := u.db.GetContext(ctx, &user, query, email, password); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return domain.User{}, domain.ErrUserNotFound
        }
        return domain.User{}, fmt.Errorf("repository: failed to get user: %w", err)
	}
	return user, nil
}

func (u *UserOperationsPostgres) SetSession(ctx context.Context, userId int64, session tokenmanager.Session) error {
	// Используем ExecContext, так как нам не нужно возвращать данные, только проверить ошибку
	// ON CONFLICT (user_id) означает: "Если запись с таким user_id уже есть..."
	// DO UPDATE SET ... означает: "...то обнови поля refresh_token и expires_in"
	
	query := fmt.Sprintf(`
		INSERT INTO %s (user_id, refresh_token, expires_in) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			refresh_token = EXCLUDED.refresh_token, 
			expires_in = EXCLUDED.expires_in
	`, refreshSessionsTable)

	_, err := u.db.ExecContext(ctx, query, userId, session.RefreshToken, session.ExpiresAt)
	
	if err != nil {
		return fmt.Errorf("failed to set session: %w", domain.ErrInsertToken)
	}

	return nil
}

func (u *UserOperationsPostgres) IsTokenValid(ctx context.Context, refreshToken string) (int64, error) {
	var userID int64
	now := time.Now()
	query := fmt.Sprintf("SELECT r.user_id FROM %s r JOIN %s u ON u.id = r.user_id WHERE r.refresh_token=$1 AND r.expires_in>$2", refreshSessionsTable, usersTable)

	row := u.db.QueryRowContext(ctx, query, refreshToken, now)

	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrInvalidToken
		}
		return 0, fmt.Errorf("db error: %w", err)
	}
	return userID, nil
}