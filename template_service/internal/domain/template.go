package domain

import "time"

type Template struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Vars        string    `db:"vars"` // Переменные шаблона в JSON-формате
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
