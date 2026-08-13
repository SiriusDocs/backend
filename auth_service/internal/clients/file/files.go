package file

import (
	"context"
	"fmt"

	pb "github.com/SiriusDocs/protos/gen/go/file"
	"google.golang.org/grpc"
)

type Client struct {
	api pb.FileServiceClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		api: pb.NewFileServiceClient(conn),
	}
}

func (c *Client) GenerateUploadURL(ctx context.Context, key, contentType string, ttlMinutes int64) (string, error) {
	resp, err := c.api.GenerateUploadURL(ctx, &pb.GenerateUploadURLRequest{
		ObjectKey:     key,
		ContentType:   contentType,
		ExpiryMinutes: ttlMinutes,
	})
	if err != nil {
		return "", fmt.Errorf("file_service error: %w", err)
	}
	return resp.GetUrl(), nil
}

func (c *Client) GenerateDownloadURL(ctx context.Context, key string, ttlMinutes int64) (string, error) {
	resp, err := c.api.GenerateDownloadURL(ctx, &pb.GenerateDownloadURLRequest{
		ObjectKey:     key,
		ExpiryMinutes: ttlMinutes,
	})
	if err != nil {
		return "", fmt.Errorf("file_service error: %w", err)
	}
	return resp.GetUrl(), nil
}
