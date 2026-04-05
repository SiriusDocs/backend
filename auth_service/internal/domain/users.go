package domain

import "time"

type User struct {
	Id                int64     `db:"id"`
	Username          string    `db:"username"`
	Email             string    `db:"email"`
	Role              string    `db:"user_role"`
	CreationTimestamp time.Time `db:"creation_timestamp"`
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}