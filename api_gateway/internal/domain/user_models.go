package domain

import "time"

type User struct {
	UUID              string    `json:"uuid"`
	Email             string    `json:"email"`
	UserName          string    `json:"user_name"`
	CreatingTimestamp time.Time `json:"creating_timestamp"`
}