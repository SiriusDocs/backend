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
	var num int
	var token sql.NullString

	query := fmt.Sprintf("SELECT refresh_token FROM %s WHERE user_id=$1", refreshSessionsTable)
	_ = u.db.GetContext(ctx, &token, query, userId)
	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		return domain.ErrTokenNotFound
	// 	}
	// }
	now := time.Now()

	if token.Valid {
		query := fmt.Sprintf("UPDATE %s SET refresh_token=$1, expires_in=$2, created_at=$3 WHERE user_id=$4 RETURNING 1", refreshSessionsTable)
		if err := u.db.GetContext(ctx, &num, query, session.RefreshToken, session.ExpiresAt, now, userId); err != nil {
			return domain.ErrInsertToken
		}
	}
	query = fmt.Sprintf("INSERT INTO %s (user_id, refresh_token, expires_in) VALUES ($1, $2, $3) RETURNING 1", refreshSessionsTable)
	if err := u.db.GetContext(ctx, &num, query, userId, session.RefreshToken, session.ExpiresAt); err != nil {
		return domain.ErrInsertToken
	}
	return nil
}

func (u *UserOperationsPostgres) IsTokenValid(ctx context.Context, refreshToken string) (int64, error) {
	var user struct {
		id       sql.NullInt64  `db:"user_id"`
		// userType sql.NullString `db:"user_type"`
	}
	now := time.Now()
	query := fmt.Sprintf("SELECT r.user_id, FROM %s r JOIN %s u ON u.id = r.user_id WHERE r.refresh_token=$1 AND r.expires_in>$2", refreshSessionsTable, usersTable)
	row := u.db.QueryRow(query, refreshToken, now)
	if err := row.Scan(&user.id,); err != nil {
		if !user.id.Valid{
			return 0, domain.ErrInvalidToken
		}
		return 0, err
	}
	return user.id.Int64, nil
}