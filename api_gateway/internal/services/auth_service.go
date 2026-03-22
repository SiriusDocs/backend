package services

import (
    "context"
    "github.com/SiriusDocs/protos/gen/go/auth"
)

type AuthService interface {
    Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error)
    Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
    GetNewTokens(ctx context.Context, req *auth.TokensRequest) (*auth.TokenResponce, error)
}

type authService struct {
    authClient auth.AuthClient
}

func NewAuthService(client auth.AuthClient) AuthService {
    return &authService{
        authClient: client,
    }
}

func (s *authService) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
    return s.authClient.Register(ctx, req)
}

func (s *authService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
    return s.authClient.Login(ctx, req)
}

func (s *authService) GetNewTokens(ctx context.Context, req *auth.TokensRequest) (*auth.TokenResponce, error) {
    return s.authClient.GetNewTokens(ctx, req)
}