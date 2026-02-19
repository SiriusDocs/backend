package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/grpc/users"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int // Порт, на котором будет работать grpc-сервер
}

func New(log *slog.Logger, service *services.Service, port int) *App {

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, // логирование тела входящего запроса
			logging.PayloadSent,     // логирование тела исходящего ответа
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
        // Логируем информацию о панике с уровнем Error
        log.Error("Recovered from panic", slog.Any("panic", p))
        
        // DEBUG лог с подробной информацией
        log.Debug("Panic details", 
            slog.Any("panic_object", p),
            slog.String("panic_type", fmt.Sprintf("%T", p)),
            slog.String("stack_trace", string(debug.Stack())),
            slog.Time("recovered_time",time.Now()),
        )
			// Можете либо честно вернуть клиенту содержимое паники
			// Либо ответить - "internal error", если не хотим делиться внутренностями
			return status.Errorf(codes.Internal, "internal error - panic")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
        recovery.UnaryServerInterceptor(recoveryOpts...), // перехват паник
        logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...), // логирование запросов/ответов
	))

	// Регистрируем наш gRPC-сервис Auth, об этом будет ниже
	users.Registered(gRPCServer, service.UserOperations)
	// регистрация остальных grpc-сервисов (из proto-контракта)
	// . . .
	// . . .

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
    return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
        l.Log(ctx, slog.Level(lvl), msg, fields...)
    })
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil{
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.grpcapp.Run"

	// Создаём listener, который будет слушать TCP-сообщения, адресованные
    // Нашему gRPC-серверу
    l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

    // Запускаем обработчик gRPC-сообщений
    if err := a.gRPCServer.Serve(l); err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }
    return nil
}

func (a *App) Stop() {
    const op = "app.grpcapp.Stop"

    a.log.With(slog.String("op", op)).
        Info("stopping gRPC server", slog.Int("port", a.port))

    // Используем встроенный в gRPCServer механизм graceful shutdown
    a.gRPCServer.GracefulStop()
}