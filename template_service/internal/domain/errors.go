package domain

import (
	"errors"
	"fmt"
)

type ErrorType int

const (
	ErrorTypeInternal   ErrorType = iota // Ошибка инфраструктуры (БД, сеть)
	ErrorTypeValidation                  // Ошибка входных данных пользователя
	ErrorTypeNotFound                    // Ресурс не найден
	ErrorTypeConflict                    // Конфликт (уже существует)
)

// AppError — единая обёртка для всех бизнес-ошибок
type AppError struct {
	Type    ErrorType
	Message string // Безопасное сообщение для пользователя
	Op      string // Операция, где произошла ошибка (для логов)
	Err     error  // Оригинальная ошибка (для логов, НИКОГДА не отдаётся клиенту)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// --- Конструкторы ---

func Validation(op, message string) *AppError {
	return &AppError{Type: ErrorTypeValidation, Message: message, Op: op}
}

func NotFound(op, message string) *AppError {
	return &AppError{Type: ErrorTypeNotFound, Message: message, Op: op}
}

func Internal(op, message string, err error) *AppError {
	return &AppError{Type: ErrorTypeInternal, Message: message, Op: op, Err: err}
}

func Conflict(op, message string) *AppError {
	return &AppError{Type: ErrorTypeConflict, Message: message, Op: op}
}

// --- Хелпер для проверки типа ---

func GetErrorType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrorTypeInternal // Неизвестная ошибка = internal
}

// GetSafeMessage возвращает сообщение, безопасное для пользователя
func GetSafeMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "internal server error"
}