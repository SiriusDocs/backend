package grpcapp

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	logs "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/pkg/logger"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// PayloadRedactInterceptor логирует request/response,
// автоматически зачищая поля с debug_redact=true
func PayloadRedactInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// ── Входящий запрос ──
		if msg, ok := req.(proto.Message); ok {
			log.DebugContext(ctx, "gRPC request payload",
                slog.String("method", info.FullMethod),
                slog.Any("body", msg),
            )
		}

		resp, err := handler(ctx, req)

		// ── Исходящий ответ ──
		if msg, ok := resp.(proto.Message); ok {
			log.DebugContext(ctx, "gRPC response payload",
                slog.String("method", info.FullMethod),
                slog.Any("body", msg),
            )
		}

		return resp, err
	}
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// TimeoutInterceptor оборачивает входящий ctx в context.WithTimeout
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if timeout <= 0 {
			return handler(ctx, req)
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}

// Один интерцептор, который собирает ВСЕ метаданные и наполняет контекст
func ContextMetadataInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		// извлекаем trace_id
		traceID := getHeader(md, "x-trace-id")
		if traceID == "" {
			traceID = uuid.New().String() // Если не передали — создаем свой
		}
		ctx = logs.WithTraceID(ctx, traceID)

		// извлекаем user_id
		if userIDStr := getHeader(md, "x-user-id"); userIDStr != "" {
			if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
				ctx = logs.WithUserID(ctx, userID)
			}
		}

		// Добавление новых заголовков в контекст по мере необходимости
		// Например, роль пользователя или IP-адрес:
		// if role := getHeader(md, "x-user-role"); role != "" {
		//     ctx = logs.WithUserRole(ctx, role)
		// }

		return handler(ctx, req)
	}
}

// Утилита для удобного получения заголовка
func getHeader(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}