package templates

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	templates "github.com/SiriusDocs/backend/template_service/internal/services"
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
	tasks  templates.TaskOperations
	params templates.ParamsOperations
	logger *slog.Logger
}

// Registered регистрирует сервер в gRPC
func Registered(gRPCServer *grpc.Server, tasks templates.TaskOperations, params templates.ParamsOperations, logger *slog.Logger) {
	pb.RegisterTempServer(gRPCServer, &TemplateServer{
		tasks:  tasks,
		params: params,
		logger: logger,
	})
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
				return status.Errorf(codes.ResourceExhausted,
					"file size exceeds limit of %d MB", maxUploadSize/(1024*1024))
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
	taskID, err := s.tasks.UploadAndStartTask(stream.Context(), filename, fileData.Bytes())
	if err != nil {
		return s.mapError(err, "UploadAndParse")
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

	fileStatus, names, err := s.tasks.CheckTaskStatus(ctx, in.TaskId)

	if err != nil {
		// ВАЖНЫЙ МОМЕНТ:
		// Если сервис вернул ошибку, но статус == "error",
		// значит это не сбой сервера, а ошибка обработки файла.
		if fileStatus == domain.TaskStatusError {
			return &pb.StatusResponse{
				Status:       fileStatus,
				Names:        nil,
				ErrorMessage: domain.GetSafeMessage(err),
			}, nil // Возвращаем nil, чтобы gRPC вызов считался успешным (HTTP 200)
		}
		return nil, s.mapError(err, "CheckStatus")
	}

	// Успех или Процесс (pending/processing/done)
	return &pb.StatusResponse{
		Status: fileStatus,
		Names:  names,
	}, nil
}

func (s *TemplateServer) CreateTemplateParams(ctx context.Context, in *pb.CreateParamsRequest) (*pb.CreateParamsResponse, error) {
	// 1. Базовая валидация входных параметров
	if in.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}
	if len(in.Params) == 0 {
		return nil, status.Error(codes.InvalidArgument, "params map is empty")
	}

	// 2. Вызов сервисного слоя (TemplateManager)
	templateID, err := s.params.CreateTemplate(ctx, in.TaskId, in.Params)
	if err != nil {
		// Здесь можно добавить более детальную обработку ошибок,
		// если сервис возвращает специфические ошибки (например, ErrTaskNotFound)
		// Но пока вернем общую ошибку с описанием
		return nil, s.mapError(err, "CreateTemplateParams")
	}

	// 3. Формирование успешного ответа
	return &pb.CreateParamsResponse{
		TemplateId: templateID,
	}, nil
}

// mapError — ЕДИНАЯ точка маппинга доменных ошибок → gRPC коды
// Логирует internal-ошибки, НЕ пробрасывает их клиенту
func (s *TemplateServer) mapError(err error, action string) error {
	errType := domain.GetErrorType(err)
	safeMsg := domain.GetSafeMessage(err)

	switch errType {
	case domain.ErrorTypeValidation:
		s.logger.Warn("validation error",
			slog.String("action", action),
			slog.String("message", safeMsg),
		)
		return status.Error(codes.InvalidArgument, safeMsg)

	case domain.ErrorTypeNotFound:
		s.logger.Warn("not found",
			slog.String("action", action),
			slog.String("message", safeMsg),
		)
		return status.Error(codes.NotFound, safeMsg)

	case domain.ErrorTypeConflict:
		s.logger.Warn("conflict",
			slog.String("action", action),
			slog.String("message", safeMsg),
		)
		return status.Error(codes.AlreadyExists, safeMsg)

	default:
		// INTERNAL: логируем полную ошибку, клиенту — безопасное сообщение
		s.logger.Error("internal error",
			slog.String("action", action),
			slog.String("error", err.Error()), // Полная ошибка — только в логи
		)
		return status.Error(codes.Internal, "internal server error")
	}
}
