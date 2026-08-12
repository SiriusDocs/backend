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
	ttl := time.Duration(req.ExpiryMinutes) * time.Minute

	url, err := s.services.GenerateDownloadURL(ctx, req.ObjectKey, ttl)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate download URL")
	}
	return &pb.GenerateDownloadURLResponse{Url: url}, nil
}

func (s *FileServer) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
	ttl := time.Duration(req.ExpiryMinutes) * time.Minute

	url, err := s.services.GenerateUploadURL(ctx, req.ObjectKey, req.ContentType, ttl)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate upload URL")
	}
	return &pb.GenerateUploadURLResponse{Url: url}, nil
}

func (s *FileServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	err := s.services.DeleteFile(ctx, req.ObjectKey)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete file")
	}
	return &pb.DeleteFileResponse{Success: true}, nil
}

func (s *FileServer) GetFileMetadata(ctx context.Context, req *pb.GetFileMetadataRequest) (*pb.GetFileMetadataResponse, error) {
	metadata, err := s.services.GetFileMetadata(ctx, req.ObjectKey)
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
