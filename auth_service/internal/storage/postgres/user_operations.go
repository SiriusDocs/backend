package postgres

import (
	"errors"
	"fmt"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
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

func (u *UserOperationsPostgres) CreateUser(username string, email string, password string) (int64, error) {
	var id int64
	query := fmt.Sprintf("INSERT INTO %s (username, email, hashed_password) values ($1, $2, $3) RETURNING id", usersTable)

	if err := u.db.Get(&id, query, username, email, password); err != nil {
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
