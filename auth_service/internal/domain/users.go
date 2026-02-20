package domain

import "time"

type User struct {
	Id                int64
	Username          string
	Email             string
	Creation_timestamp time.Time
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
