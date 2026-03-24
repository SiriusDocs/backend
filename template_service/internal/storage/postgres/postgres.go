package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	tasksTable = "processing_tasks"
)

func NewPostgresDB(host string, port string, username string, DBName string, password string, sslMode string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", host, port, username, DBName, password, sslMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}