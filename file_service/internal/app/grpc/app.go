package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/grpc/files"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/file_service/internal/services"
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

func New(log *slog.Logger, service *services.Service, port int, timeout time.Duration) *App {

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.StartCall,  // ← только метаданные вызова
			logging.FinishCall, // ← результат (код, время)
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
				slog.Time("recovered_time", time.Now()),
			)
			// Можете либо честно вернуть клиенту содержимое паники
			// Либо ответить - "internal error", если не хотим делиться внутренностями
			return status.Errorf(codes.Internal, "internal error - panic")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		ContextMetadataInterceptor(),                                           // добавление метаданных в контекст
		TimeoutInterceptor(timeout),                                            // таймаут на обработку запроса
		recovery.UnaryServerInterceptor(recoveryOpts...),                       // перехват паник
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...), // логирование запросов/ответов
		PayloadRedactInterceptor(log),                                          // безопасный логгер payload'ов
	))

	files.Registered(gRPCServer, service.FileOperations)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
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
	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	stopped := make(chan struct{})
	go func() {
		a.gRPCServer.GracefulStop()
		close(stopped)
	}()

	// Даем 5 секунд на мягкое завершение, иначе выключаем жестко
	select {
	case <-stopped:
		a.log.Info("gRPC server stopped gracefully")
	case <-time.After(5 * time.Second):
		a.log.Warn("gRPC server timeout reached, forcing shutdown")
		a.gRPCServer.Stop()
	}
}
