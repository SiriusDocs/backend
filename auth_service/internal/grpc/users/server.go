package users

import (
	"context"
	"errors"

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
}

func Registered(gRPCServer *grpc.Server, Service services.UserOperations) {
	pb.RegisterAuthServer(gRPCServer, &UsersServer{services: Service})
}

func (u *UsersServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Username == "" || in.Email == "" || in.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "all fields are required")
    }

	id, err := u.services.CreateUser(in.Username, in.Email, in.Password)
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