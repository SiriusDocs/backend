package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Storage struct {
	client *s3.Client
	bucket string
}

func New(ctx context.Context, cfg config.S3) (*Storage, error) {
	awsConf, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.Region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsConf, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *Storage) PresignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, error) {
	ps := s3.NewPresignClient(s.client)
	req, err := ps.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return "", fmt.Errorf("s3 presign put %q: %w", key, err)
	}

	return req.URL, nil
}

func (s *Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	ps := s3.NewPresignClient(s.client)
	req, err := ps.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return "", fmt.Errorf("s3 presign get %q: %w", key, err)
	}

	return req.URL, nil
}

func (s *Storage) Head(ctx context.Context, key string) (domain.FileMetadata, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			// Ошибку сети или прав доступа не кидаем, просто говорим: "Файла нет"
			return domain.FileMetadata{
				Exists: false,
			}, nil
		}
		
		// Если это другая ошибка (нет интернета, неверный токен и т.д.) - возвращаем ее
		return domain.FileMetadata{Exists: false}, fmt.Errorf("s3 head %q: %w", key, err)
	}
	
	// Безопасно достаем время и переводим в строку
	var lastModStr string
	if out.LastModified != nil {
		lastModStr = out.LastModified.Format(time.RFC3339) 
	}
	
	return domain.FileMetadata{
		Exists:       true,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		LastModified: lastModStr,
	}, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("s3 delete %q: %w", key, err)
	}

	return nil
}
