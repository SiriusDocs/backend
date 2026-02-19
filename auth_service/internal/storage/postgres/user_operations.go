package postgres

import (
	"github.com/jmoiron/sqlx"
)

type UserOperationsPostgres struct{
	db *sqlx.DB
}

func NewUserOperationsPostgres(db *sqlx.DB) *UserOperationsPostgres {
	return &UserOperationsPostgres{
		db: db,
	}
}