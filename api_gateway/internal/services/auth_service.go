package services

import (
    "context"
    "github.com/SiriusDocs/protos/gen/go/auth"
)

type AuthService interface {
    Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error)
    Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
    GetNewTokens(ctx context.Context, req *auth.TokensRequest) (*auth.TokenResponse, error)
    GetProfile(ctx context.Context, req *auth.GetProfileRequest) (*auth.GetProfileResponse, error)
	ListPendingUsers(ctx context.Context, req *auth.ListPendingUsersRequest) (*auth.ListPendingUsersResponse, error)
	AssignRole(ctx context.Context, req *auth.AssignRoleRequest) (*auth.AssignRoleResponse, error)
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

func (s *authService) GetNewTokens(ctx context.Context, req *auth.TokensRequest) (*auth.TokenResponse, error) {
    return s.authClient.GetNewTokens(ctx, req)
}

func (s *authService) GetProfile(ctx context.Context, req *auth.GetProfileRequest) (*auth.GetProfileResponse, error) {
	return s.authClient.GetProfile(ctx, req)
}

func (s *authService) ListPendingUsers(ctx context.Context, req *auth.ListPendingUsersRequest) (*auth.ListPendingUsersResponse, error) {
	return s.authClient.ListPendingUsers(ctx, req)
}

func (s *authService) AssignRole(ctx context.Context, req *auth.AssignRoleRequest) (*auth.AssignRoleResponse, error) {
	return s.authClient.AssignRole(ctx, req)
}