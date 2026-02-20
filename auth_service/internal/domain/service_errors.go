package domain

import "errors"

var (
    ErrUserExists   = errors.New("user already exists")
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidCreds = errors.New("invalid credentials")
)