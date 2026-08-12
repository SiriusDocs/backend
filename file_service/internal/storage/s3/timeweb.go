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
	const op = "storage.s3.New"

	awsConf, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.Region),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: load aws config: %w", op, err)
	}

	client := s3.NewFromConfig(awsConf, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *Storage) PresignPut(ctx context.Context, key, contentType string, ttl time.Duration) (string, error) {
	const op = "storage.s3.PresignPut"

	ps := s3.NewPresignClient(s.client)
	req, err := ps.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return "", fmt.Errorf("%s (key=%s): %w", op, key, err)
	}

	return req.URL, nil
}

func (s *Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	const op = "storage.s3.PresignGet"

	ps := s3.NewPresignClient(s.client)
	req, err := ps.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))

	if err != nil {
		return "", fmt.Errorf("%s (key=%s): %w", op, key, err)
	}

	return req.URL, nil
}

func (s *Storage) Head(ctx context.Context, key string) (domain.FileMetadata, error) {
	const op = "storage.s3.Head"

	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		// В AWS SDK v2 ненайденный объект может возвращать NotFound или NoSuchKey
		var notFound *types.NotFound
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &notFound) || errors.As(err, &noSuchKey) {
			// Отсутствие файла — это не ошибка сети или S3, это корректный бизнес-ответ
			return domain.FileMetadata{
				Exists: false,
			}, nil
		}

		// Любая другая ошибка (нет сети, 403 Forbidden, 500 S3 Error) оборачивается
		return domain.FileMetadata{Exists: false}, fmt.Errorf("%s (key=%s): %w", op, key, err)
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
	const op = "storage.s3.Delete"

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("%s (key=%s): %w", op, key, err)
	}

	return nil
}
