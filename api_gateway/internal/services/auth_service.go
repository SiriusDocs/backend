package services

import (
    "context"
    "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/api_gateway/pkg/api/auth"
)

// Интерфейс остаётся тем же
type AuthService interface {
    Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error)
    Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
}

type authService struct {
    authClient auth.AuthClient // gRPC сгенерированный клиент
}

// Изменение здесь: мы передаём УЖЕ готовый клиент
func NewAuthService(client auth.AuthClient) AuthService {
    return &authService{
        authClient: client,
    }
}

func (s *authService) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
    // Убрали проверку на nil, так как клиент теперь гарантированно передается при создании
    return s.authClient.Register(ctx, req)
}

func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
    return s.authClient.Login(ctx, req)
}