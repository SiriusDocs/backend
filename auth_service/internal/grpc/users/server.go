package users

import (
	"context"

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
	AddUser(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) 
}

func Registered(gRPCServer *grpc.Server, Service services.UserOperations) {
	pb.RegisterAuthServer(gRPCServer, &UsersServer{services: Service})
}

func (u *UsersServer) AddUser(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

}