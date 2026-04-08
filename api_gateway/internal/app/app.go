package app

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	client "github.com/SiriusDocs/backend/api_gateway/internal/app/grpc"
	httpapp "github.com/SiriusDocs/backend/api_gateway/internal/app/http"
	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/SiriusDocs/backend/api_gateway/internal/transport"
	authHandler "github.com/SiriusDocs/backend/api_gateway/internal/transport/handlers/auth"
	tempHandler "github.com/SiriusDocs/backend/api_gateway/internal/transport/handlers/templates"
	"github.com/SiriusDocs/protos/gen/go/auth"
	"github.com/SiriusDocs/protos/gen/go/templates"
)

type App struct {
	HTTPServer *httpapp.App
	grpcConns  []*grpc.ClientConn
	log        *slog.Logger
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	app := &App{
		log:       log,
		grpcConns: make([]*grpc.ClientConn, 0),
	}

	authConn, err := app.connectService("auth-service", cfg.Clients.AuthService.Address)
	if err != nil {
		return nil, err
	}

	templateConn, err := app.connectService("template-service", cfg.Clients.TempService.Address)
	if err != nil {
		return nil, err
	}

	// documentConn, err := grpcclient.NewClient(cfg.Clients.DocumentService.Address, log)
	// if err != nil { panic(err) }
	// conns = append(conns, documentConn)

	authClient := auth.NewAuthClient(authConn)
	authService := services.NewAuthService(authClient)

	templateClient := templates.NewTempClient(templateConn)
	templateService := services.NewTempService(templateClient)

	// documentClient := billing.NewDocumentClient(documentConn)
	// documentService := services.NewDocumentService(documentClient)

	router := transport.NewRouter(cfg)
	apiGroup := router.Group("/")

	authHndlr := authHandler.New(log, authService, &cfg.Clients.AuthService, cfg.JWTSecret)
	authHndlr.RegisterRoutes(apiGroup)

	tempHndlr := tempHandler.New(log, templateService, &cfg.Clients.TempService)
	tempHndlr.RegisterRoutes(apiGroup)
	// docHndlr := docHandler.New(log, documentService)
	// docHndlr.RegisterRoutes(apiGroup)

	app.HTTPServer = httpapp.New(log, cfg.HTTPServer.Port, router)

	return app, nil
}


// connectService подключается к gRPC сервису с логированием
func (a *App) connectService(name, address string) (*grpc.ClientConn, error) {
	conn, err := client.NewClient(address, a.log)
	if err != nil {
		a.log.Error("failed to connect to grpc service",
			slog.String("service", name),
			slog.String("address", address),
			slog.String("error", err.Error()),
		)

		// Закрываем уже открытые соединения
		a.Close()

		return nil, fmt.Errorf("failed to connect to %s at %s: %w", name, address, err)
	}

	a.log.Info("connected to grpc service",
		slog.String("service", name),
		slog.String("address", address),
	)

	a.grpcConns = append(a.grpcConns, conn)
	return conn, nil
}

// Close закрывает все gRPC соединения
func (a *App) Close() {
	for _, conn := range a.grpcConns {
		if conn != nil {
			conn.Close()
		}
	}
	a.log.Info("all grpc connections closed")
}