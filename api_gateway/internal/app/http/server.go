package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	log        *slog.Logger
	httpServer *http.Server
	port       int
}

func New(log *slog.Logger, port int, handler http.Handler) *App {
	return &App{
		log:  log,
		port: port,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"
	a.log.Info("http server started", slog.String("addr", a.httpServer.Addr))

	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "httpapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping http server", slog.Int("port", a.port))

	return a.httpServer.Shutdown(ctx)
}
