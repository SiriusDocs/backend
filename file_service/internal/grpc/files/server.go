package files

import (
	"context"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/services"
	pb "github.com/SiriusDocs/protos/gen/go/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: закинуть в конфиг
const defaultExpiryMinutes = 15 // Значение по умолчанию для TTL ссылки

type FileServer struct {
	pb.UnimplementedFileServiceServer
	services services.FileOperations
}

type FilesServer interface {
	GenerateDownloadURL(ctx context.Context, req *pb.GenerateDownloadURLRequest) (*pb.GenerateDownloadURLResponse, error)
	GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error)
	DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error)
	GetFileMetadata(ctx context.Context, req *pb.GetFileMetadataRequest) (*pb.GetFileMetadataResponse, error)
	// CopyFile(ctx context.Context, req *pb.CopyFileRequest) (*pb.CopyFileResponse, error)
}

func Registered(gRPCServer *grpc.Server, Service services.FileOperations) {
	pb.RegisterFileServiceServer(gRPCServer, &FileServer{services: Service})
}

func (s *FileServer) GenerateDownloadURL(ctx context.Context, req *pb.GenerateDownloadURLRequest) (*pb.GenerateDownloadURLResponse, error) {
	if req.GetObjectKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "object_key is required")
	}

	expiryMinutes := req.GetExpiryMinutes()
	if expiryMinutes <= 0 {
		expiryMinutes = defaultExpiryMinutes
	}

	ttl := time.Duration(expiryMinutes) * time.Minute

	url, err := s.services.GenerateDownloadURL(ctx, req.GetObjectKey(), ttl)
	if err != nil {
		// ошибка уже залогирована внутри service с полным стек-трейсом
		// клиенту отдаем понятный статус без слива внутренних деталей
		return nil, status.Error(codes.Internal, "failed to generate download URL")
	}

	return &pb.GenerateDownloadURLResponse{Url: url}, nil
}

func (s *FileServer) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
	if req.GetObjectKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "object_key is required")
	}
	if req.GetContentType() == "" {
		return nil, status.Error(codes.InvalidArgument, "content_type is required")
	}

	expiryMinutes := req.GetExpiryMinutes()
	if expiryMinutes <= 0 {
		expiryMinutes = defaultExpiryMinutes
	}

	ttl := time.Duration(expiryMinutes) * time.Minute

	url, err := s.services.GenerateUploadURL(ctx, req.GetObjectKey(), req.GetContentType(), ttl)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate upload URL")
	}

	return &pb.GenerateUploadURLResponse{Url: url}, nil
}

func (s *FileServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	if req.GetObjectKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "object_key is required")
	}

	err := s.services.DeleteFile(ctx, req.GetObjectKey())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete file")
	}

	return &pb.DeleteFileResponse{Success: true}, nil
}

func (s *FileServer) GetFileMetadata(ctx context.Context, req *pb.GetFileMetadataRequest) (*pb.GetFileMetadataResponse, error) {
	if req.GetObjectKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "object_key is required")
	}

	metadata, err := s.services.GetFileMetadata(ctx, req.GetObjectKey())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get file metadata")
	}

	return &pb.GetFileMetadataResponse{
		Exists:       metadata.Exists,
		ContentType:  metadata.ContentType,
		LastModified: metadata.LastModified,
	}, nil
}
// func (s *FileServer) CopyFile(ctx context.Context, req *pb.CopyFileRequest) (*pb.CopyFileResponse, error)
