package grpcapp

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

// redactMessage клонирует сообщение и зачищает все поля с debug_redact=true
func redactMessage(msg proto.Message) proto.Message {
	if msg == nil {
		return nil
	}
	cloned := proto.Clone(msg)
	redactFields(cloned.ProtoReflect())
	return cloned
}

func redactFields(m protoreflect.Message) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		// ── Проверяем debug_redact ──
		if opts, ok := fd.Options().(*descriptorpb.FieldOptions); ok &&
			opts != nil && opts.GetDebugRedact() {

			switch fd.Kind() {
			case protoreflect.StringKind:
				m.Set(fd, protoreflect.ValueOfString("[REDACTED]"))
			case protoreflect.BytesKind:
				m.Set(fd, protoreflect.ValueOfBytes([]byte("[REDACTED]")))
			default:
				m.Clear(fd) // числа, bool и т.д. — просто обнуляем
			}
			return true
		}

		// ── Рекурсия во вложенные message ──
		if fd.Kind() == protoreflect.MessageKind {
			switch {
			case fd.IsList():
				list := v.List()
				for i := 0; i < list.Len(); i++ {
					redactFields(list.Get(i).Message())
				}
			case fd.IsMap():
				v.Map().Range(func(_ protoreflect.MapKey, val protoreflect.Value) bool {
					if fd.MapValue().Kind() == protoreflect.MessageKind {
						redactFields(val.Message())
					}
					return true
				})
			default:
				redactFields(v.Message())
			}
		}

		return true
	})
}

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
				slog.Any("body", redactMessage(msg)),
			)
		}

		resp, err := handler(ctx, req)

		// ── Исходящий ответ ──
		if msg, ok := resp.(proto.Message); ok {
			log.InfoContext(ctx, "response payload",
				slog.String("method", info.FullMethod),
				slog.Any("body", redactMessage(msg)),
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

