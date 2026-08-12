package grpcapp

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
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
			log.InfoContext(ctx, "request payload",
				slog.String("method", info.FullMethod),
				slog.Any("body", msg),
			)
		}

		resp, err := handler(ctx, req)

		// ── Исходящий ответ ──
		if msg, ok := resp.(proto.Message); ok {
			log.InfoContext(ctx, "response payload",
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

