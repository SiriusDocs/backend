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
		fmt.Println(err)
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
