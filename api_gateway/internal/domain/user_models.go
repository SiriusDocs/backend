package domain

import "time"

type User struct {
	UUID              string    `json:"uuid"`
	Email             string    `json:"email"`
	UserName          string    `json:"user_name"`
	CreatingTimestamp time.Time `json:"creating_timestamp"`
}

type RegisterRequest struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
