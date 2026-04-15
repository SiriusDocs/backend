package services

import (
	"context"
	"io"

	"github.com/SiriusDocs/protos/gen/go/file"
)

type FileService interface {
	// Upload
	UploadFile(ctx context.Context, meta *files.UploadMeta, fileReader io.Reader) (*files.UploadFileResponse, error)
	
	// Download
	DownloadFile(ctx context.Context, fileID, userID string, roles []string) (io.ReadCloser, error)
	
	// Progress
	GetUploadStatus(ctx context.Context, uploadID string) (int32, bool, error)
	
	// File Info
	GetFileMeta(ctx context.Context, fileID string) (*files.FileMeta, error)
	
	// Admin
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteFile(ctx context.Context, fileID string) error
	ListFiles(ctx context.Context, bucket string) ([]*files.FileMeta, error)
}

type fileService struct {
	fileClient files.FileClient // <- Правильное имя для сервиса "File"
}

func NewFileService(client files.FileClient) FileService {
	return &fileService{
		fileClient: client,
	}
}

// UploadFile загружает файл через stream
func (s *fileService) UploadFile(ctx context.Context, meta *files.UploadMeta, fileReader io.Reader) (*files.UploadFileResponse, error) {
	stream, err := s.fileClient.UploadFile(ctx)
	if err != nil {
		return nil, err
	}

	// Отправляем метаданные (первое сообщение)
	err = stream.Send(&files.UploadFileRequest{
		Data: &files.UploadFileRequest_Meta{
			Meta: meta,
		},
	})
	if err != nil {
		return nil, err
	}

	// Читаем файл кусками и отправляем в стрим
	buffer := make([]byte, 32*1024) // 32KB chunks
	for {
		n, err := fileReader.Read(buffer)
		if err == io.EOF {
			break // конец файла
		}
		if err != nil {
			return nil, err
		}

		err = stream.Send(&files.UploadFileRequest{
			Data: &files.UploadFileRequest_Chunk{
				Chunk: &files.FileChunk{
					Content: buffer[:n],
				},
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// Закрываем стрим и получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DownloadFile скачивает файл через stream и возвращает io.ReadCloser
func (s *fileService) DownloadFile(ctx context.Context, fileID, userID string, roles []string) (io.ReadCloser, error) {
	stream, err := s.fileClient.DownloadFile(ctx, &files.DownloadFileRequest{
		FileId: fileID,
		UserId: userID,
		Roles:  roles,
	})
	if err != nil {
		return nil, err
	}

	// Возвращаем reader, который будет читать из stream
	return newStreamReader(stream), nil
}

// GetUploadStatus проверяет прогресс загрузки
func (s *fileService) GetUploadStatus(ctx context.Context, uploadID string) (int32, bool, error) {
	resp, err := s.fileClient.GetUploadStatus(ctx, &files.GetUploadStatusRequest{
		UploadId: uploadID,
	})
	if err != nil {
		return 0, false, err
	}

	return resp.ProgressPercent, resp.Completed, nil
}

// GetFileMeta получает метаданные файла
func (s *fileService) GetFileMeta(ctx context.Context, fileID string) (*files.FileMeta, error) {
	resp, err := s.fileClient.GetFileMeta(ctx, &files.GetFileMetaRequest{
		FileId: fileID,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateBucket создает новый bucket (admin)
func (s *fileService) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := s.fileClient.CreateBucket(ctx, &files.CreateBucketRequest{
		BucketName: bucketName,
	})
	return err
}

// DeleteFile удаляет файл (admin)
func (s *fileService) DeleteFile(ctx context.Context, fileID string) error {
	_, err := s.fileClient.DeleteFile(ctx, &files.DeleteFileRequest{
		FileId: fileID,
	})
	return err
}

// ListFiles получает список файлов в bucket
func (s *fileService) ListFiles(ctx context.Context, bucket string) ([]*files.FileMeta, error) {
	resp, err := s.fileClient.ListFiles(ctx, &files.ListFilesRequest{
		Bucket: bucket,
	})
	if err != nil {
		return nil, err
	}

	return resp.Files, nil
}

// ========================================
// streamReader - обёртка для stream в io.ReadCloser
// ========================================

type streamReader struct {
	stream files.File_DownloadFileClient // <- Правильный тип для сервиса "File"
	buffer []byte
}

func newStreamReader(stream files.File_DownloadFileClient) *streamReader {
	return &streamReader{
		stream: stream,
		buffer: []byte{},
	}
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	// Если буфер пуст, получаем следующий chunk
	if len(r.buffer) == 0 {
		chunk, err := r.stream.Recv()
		if err != nil {
			return 0, err // EOF или ошибка
		}
		r.buffer = chunk.Chunk
	}

	// Копируем из буфера в p
	n = copy(p, r.buffer)
	r.buffer = r.buffer[n:]

	return n, nil
}

func (r *streamReader) Close() error {
	return r.stream.CloseSend()
}