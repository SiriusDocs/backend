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
    // В современном gRPC таймауты задаются при самом вызове метода (в хэндлере), 
    // а не при создании соединения.
    
    opts := []grpc.DialOption{
        // Пока используем insecure. В проде тут будут TLS сертификаты
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithChainUnaryInterceptor(
            loggingInterceptor(log),
        ),
    }

    conn, err := grpc.NewClient(addr, opts...)
    if err != nil {
        return nil, fmt.Errorf("failed to create gRPC client for %s: %w", addr, err)
    }

    return conn, nil
}