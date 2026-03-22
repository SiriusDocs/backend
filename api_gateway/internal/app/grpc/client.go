package client

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// loggingInterceptor логирует все вызовы к микросервисам
func loggingInterceptor(log *slog.Logger) grpc.UnaryClientInterceptor {
    return func(
        ctx context.Context, method string, req, reply interface{},
        cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
    ) error {
        start := time.Now()
        
        err := invoker(ctx, method, req, reply, cc, opts...)
        
        log.Debug("gRPC call", 
            slog.String("method", method),
            slog.Duration("duration", time.Since(start)),
            slog.Any("error", err),
        )
        return err
    }
}

// NewClient — универсальная фабрика для создания gRPC-соединений
func NewClient(addr string, log *slog.Logger) (*grpc.ClientConn, error) {
	// Проверка на пустой адрес
	if addr == "" {
		return nil, fmt.Errorf("grpc address is empty")
	}

	// Таймаут на подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			loggingInterceptor(log),
		),
		grpc.WithBlock(), // Ждём реального подключения
	}

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return conn, nil
}