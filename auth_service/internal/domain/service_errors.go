package domain

import "errors"

var (
    ErrUserExists   = errors.New("user already exists")
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidCreds = errors.New("invalid credentials")

    ErrTokenNotFound = errors.New("refresh token not found")
    ErrInsertToken = errors.New("the new token cannot be written to the database")
    ErrInvalidToken = errors.New("invalid refresh token")
)