package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// level — общий порог логирования. LevelVar потокобезопасен, поэтому уровень
// можно менять в рантайме (например, из админской ручки) без перезапуска сервиса.
var level slog.LevelVar

// SetLevel меняет порог логирования на лету.
func SetLevel(l slog.Level) { level.Set(l) }

func SetupLogger(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case envDev:
		level.Set(slog.LevelDebug)
		handler = slog.NewJSONHandler(os.Stdout, options(true))
	case envProd:
		level.Set(slog.LevelInfo)
		handler = slog.NewJSONHandler(os.Stdout, options(true))
	case envLocal:
		fallthrough
	default:
		level.Set(slog.LevelDebug)
		handler = slog.NewTextHandler(os.Stdout, options(false))
	}

	//return slog.New(ContextHandler{Handler: handler})
	return slog.New(handler)
}

func options(addSource bool) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:       &level,
		AddSource:   addSource,
		ReplaceAttr: replaceAttr,
	}
}

// replaceAttr приводит запись к единому виду:
// время в UTC/RFC3339, источник — коротким
// путём file:line.
// Плюс страховочная сетка на секреты: основная защита — LogValue()
// на самих типах, здесь ловим то, что залогировали по имени поля мимо неё.
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	// преобразует время в UTC/RFC3339Nano
	case slog.TimeKey:
		if len(groups) == 0 {
			a.Value = slog.StringValue(a.Value.Time().UTC().Format(time.RFC3339Nano))
		}
	// укорачивает запись источника до file:line
	case slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok {
			a.Value = slog.StringValue(filepath.Base(src.File) + ":" + strconv.Itoa(src.Line))
		}
	case "password", "token", "secret", "secret_key", "access_key", "authorization":
		a.Value = slog.StringValue("[REDACTED]")
	}

	return a
}

// TODO: для связки с другими микросервисами

// type ctxKey struct{ name string }

// var (
// 	traceIDKey = ctxKey{"trace_id"}
// 	userIDKey  = ctxKey{"user_id"}
// )

// // WithTraceID кладёт сквозной идентификатор запроса в контекст.
// // Ставится на входе в сервис (gRPC-интерсептором из метаданных api_gateway).
// func WithTraceID(ctx context.Context, traceID string) context.Context {
// 	return context.WithValue(ctx, traceIDKey, traceID)
// }

// // WithUserID кладёт идентификатор пользователя в контекст.
// func WithUserID(ctx context.Context, userID int64) context.Context {
// 	return context.WithValue(ctx, userIDKey, userID)
// }

// // ContextHandler дописывает в каждую запись набор request-полей из контекста,
// // чтобы их не приходилось передавать руками в каждый вызов логгера.
// // Работает только с методами вида InfoContext/ErrorContext — им передаётся ctx.
// type ContextHandler struct {
// 	slog.Handler
// }

// func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
// 	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
// 		r.AddAttrs(slog.String("trace_id", traceID))
// 	}
// 	if userID, ok := ctx.Value(userIDKey).(int64); ok {
// 		r.AddAttrs(slog.Int64("user_id", userID))
// 	}

// 	return h.Handler.Handle(ctx, r)
// }

// // WithAttrs и WithGroup обязаны сохранять обёртку, иначе log.With(...) вернёт
// // «голый» хендлер и поля из контекста перестанут добавляться.
// func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
// 	return ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
// }

// func (h ContextHandler) WithGroup(name string) slog.Handler {
// 	return ContextHandler{Handler: h.Handler.WithGroup(name)}
// }
