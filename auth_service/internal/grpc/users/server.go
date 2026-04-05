package users

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	user, err := u.services.GetProfile(ctx, in.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return &pb.GetProfileResponse{
		UserId:   user.Id,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (u *UsersServer) ListPendingUsers(ctx context.Context, in *pb.ListPendingUsersRequest) (*pb.ListPendingUsersResponse, error) {
	limit := in.Limit
	if limit == 0 {
		limit = 10 // дефолт
	}

	users, total, err := u.services.GetPendingUsers(ctx, limit, in.Offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	var pbUsers []*pb.PendingUser
	for _, user := range users {
		pbUsers = append(pbUsers, &pb.PendingUser{
			UserId:    user.Id,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreationTimestamp.Format(time.RFC3339),
		})
	}

	return &pb.ListPendingUsersResponse{
		Users:      pbUsers,
		TotalCount: total,
	}, nil
}

func (u *UsersServer) AssignRole(ctx context.Context, in *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	if in.TargetUserId == 0 || in.NewRole == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	err := u.services.AssignRole(ctx, in.TargetUserId, in.NewRole)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to assign role")
	}

	return &pb.AssignRoleResponse{Success: true}, nil
}