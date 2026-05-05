package fill

import (
	// "bytes"
	// "context"
	// "io"
	"log/slog"

	"github.com/SiriusDocs/backend/fill_service/internal/domain"
	fill "github.com/SiriusDocs/backend/fill_service/internal/services"
	pb "github.com/SiriusDocs/protos/gen/go/templates"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FillServer struct {
	pb.UnimplementedTempServer
	services *fill.Service
	logger   *slog.Logger
}

// Registered регистрирует сервер в gRPC
func Registered(gRPCServer *grpc.Server, services *fill.Service, logger *slog.Logger) {
	pb.RegisterTempServer(gRPCServer, &FillServer{
		services: services,
		logger:   logger,
	})
}

// mapError — ЕДИНАЯ точка маппинга доменных ошибок → gRPC коды
// Логирует internal-ошибки, НЕ пробрасывает их клиенту
func (s *FillServer) mapError(err error, action string) error {
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
