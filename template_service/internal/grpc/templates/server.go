package templates

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/SiriusDocs/backend/template_service/internal/services"
	pb "github.com/SiriusDocs/protos/gen/go/templates"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Лимит на размер буфера в памяти при приеме стрима (чтобы не забить RAM до валидации)
// Должен быть >= MaxFileSize в сервисе.
const maxUploadSize = 20 * 1024 * 1024 // 20 MB

type TemplateServer struct {
	pb.UnimplementedTempServer
	service templates.TaskOperations
}

// Registered регистрирует сервер в gRPC
func Registered(gRPCServer *grpc.Server, service templates.TaskOperations) {
	pb.RegisterTempServer(gRPCServer, &TemplateServer{service: service})
}

// UploadAndParse принимает поток данных (файл), собирает его и передает в сервис
func (s *TemplateServer) UploadAndParse(stream pb.Temp_UploadAndParseServer) error {
	var (
		fileData bytes.Buffer
		filename string
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Конец стрима, выходим из цикла
			break
		}
		if err != nil {
			return status.Error(codes.Unknown, "failed to receive stream chunk")
		}

		// Обрабатываем oneof поля (метаданные или чанк файла)
		switch data := req.Data.(type) {
		case *pb.UploadRequest_Info:
			filename = data.Info.Filename

		case *pb.UploadRequest_ChunkData:
			// Защита от переполнения памяти (DoS)
			if fileData.Len()+len(data.ChunkData) > maxUploadSize {
				return status.Errorf(codes.ResourceExhausted, "file size exceeds limit of %d bytes", maxUploadSize)
			}
			fileData.Write(data.ChunkData)
		}
	}

	if filename == "" {
		return status.Error(codes.InvalidArgument, "filename metadata is missing")
	}

	if fileData.Len() == 0 {
		return status.Error(codes.InvalidArgument, "file content is empty")
	}

	// Передаем собранный файл в сервисный слой
	// Контекст берем из стрима
	taskID, err := s.service.UploadAndStartTask(stream.Context(), filename, fileData.Bytes())
	if err != nil {
		// Сервис возвращает валидационные ошибки.
		// В идеале можно проверять тип ошибки, но пока вернем Internal или InvalidArgument
		// Если ошибка "validation failed", можно считать это ошибкой клиента
		return status.Errorf(codes.Internal, "failed to process file: %v", err)
	}

	// Отправляем ответ и закрываем стрим
	return stream.SendAndClose(&pb.UploadResponse{
		TaskId: taskID,
	})
}

// CheckStatus проверяет статус задачи
func (s *TemplateServer) CheckStatus(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	if in.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	fileStatus, names, err := s.service.CheckTaskStatus(ctx, in.TaskId)
	
	if err != nil {
		// ВАЖНЫЙ МОМЕНТ:
		// Если сервис вернул ошибку, но статус == "error", 
		// значит это не сбой сервера, а ошибка обработки файла.
		if fileStatus == domain.TaskStatusError {
			return &pb.StatusResponse{
				Status:       fileStatus, // "error"
				Names:        nil,
				ErrorMessage: err.Error(), // Текст ошибки из сервиса ("validation failed" и т.д.)
			}, nil // Возвращаем nil, чтобы gRPC вызов считался успешным (HTTP 200)
		}

		// Если статус пустой или другой, значит это реальная ошибка БД или кода
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}
		
		return nil, status.Errorf(codes.Internal, "failed to check status: %v", err)
	}

	// Успех или Процесс (pending/processing/done)
	return &pb.StatusResponse{
		Status: fileStatus,
		Names:  names,
	}, nil
}