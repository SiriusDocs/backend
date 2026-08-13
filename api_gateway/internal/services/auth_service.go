package services

import (
    "context"
    "github.com/SiriusDocs/protos/gen/go/auth"
)

type AuthService interface {
	Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error)
	Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error)
	GetNewTokens(ctx context.Context, req *auth.TokensRequest) (*auth.TokenResponse, error)
	
	// Profile & Avatar methods
	GetProfile(ctx context.Context, req *auth.GetProfileRequest) (*auth.GetProfileResponse, error)
	GetAvatar(ctx context.Context, req *auth.GetAvatarRequest) (*auth.GetAvatarResponse, error)
	GenerateAvatarUploadUrl(ctx context.Context, req *auth.GenerateAvatarUploadUrlRequest) (*auth.GenerateAvatarUploadUrlResponse, error)
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

func (s *authService) GetAvatar(ctx context.Context, req *auth.GetAvatarRequest) (*auth.GetAvatarResponse, error) {
	return s.authClient.GetAvatar(ctx, req)
}

func (s *authService) GenerateAvatarUploadUrl(ctx context.Context, req *auth.GenerateAvatarUploadUrlRequest) (*auth.GenerateAvatarUploadUrlResponse, error) {
	return s.authClient.GenerateAvatarUploadUrl(ctx, req)
}