package domain

import (
	"time"
)

const (
	DocStatusDraft     = "draft"
	DocStatusReview    = "on_review"
	DocStatusFinished  = "finished"
	DocStatusPublished = "published"
)

// Соответствует таблице documents
type Document struct {
	Id        string    `db:"id"`
	Template  string    `db:"template"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
