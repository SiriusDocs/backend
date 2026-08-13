package users

import (
	"context"
	"errors"
	"fmt"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services"
	pb "github.com/SiriusDocs/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersServer struct {
	pb.UnimplementedAuthServer
	services services.UserOperations
}

type AuthServer interface {
	Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error)
	Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error)
	GetNewTokens(ctx context.Context, in *pb.TokensRequest) (*pb.TokenResponse, error)
}

func Registered(gRPCServer *grpc.Server, Service services.UserOperations) {
	pb.RegisterAuthServer(gRPCServer, &UsersServer{services: Service})
}

func (u *UsersServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}

	id, err := u.services.CreateUser(ctx, in.Username, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.RegisterResponse{
		UserId: id,
	}, nil
}

func (u *UsersServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}
	_, tokens, err := u.services.GenerateTokens(ctx, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInsertToken) || errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		} else if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (u *UsersServer) GetNewTokens(ctx context.Context, in *pb.TokensRequest) (*pb.TokenResponse, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}
	tokens, err := u.services.RefreshToken(ctx, in.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInsertToken) || errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		} else if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		fmt.Println(err)
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return &pb.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (u *UsersServer) GetProfile(ctx context.Context, in *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, avatarURL, err := u.services.GetProfile(ctx, in.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user profile")
	}

	return &pb.GetProfileResponse{
		UserId:    user.Id,
		Username:  user.Username,
		Email:     user.Email,
		UserRole:  user.Role,
		AvatarUrl: avatarURL,
	}, nil
}

func (u *UsersServer) GetAvatar(ctx context.Context, in *pb.GetAvatarRequest) (*pb.GetAvatarResponse, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	avatarURL, err := u.services.GetAvatarURL(ctx, in.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, domain.ErrAvatarNotFound) {
			return nil, status.Error(codes.NotFound, "avatar not found")
		}
		return nil, status.Error(codes.Internal, "failed to get avatar url")
	}

	return &pb.GetAvatarResponse{
		AvatarUrl: avatarURL,
	}, nil
}

func (u *UsersServer) GenerateAvatarUploadUrl(ctx context.Context, in *pb.GenerateAvatarUploadUrlRequest) (*pb.GenerateAvatarUploadUrlResponse, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if in.GetContentType() == "" {
		return nil, status.Error(codes.InvalidArgument, "content_type is required")
	}

	uploadURL, avatarKey, err := u.services.GenerateAvatarUploadURL(ctx, in.GetUserId(), in.GetContentType())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to generate avatar upload url")
	}

	return &pb.GenerateAvatarUploadUrlResponse{
		UploadUrl: uploadURL,
		AvatarKey: avatarKey,
	}, nil
}
