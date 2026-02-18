package app

import (
	"log/slog"

	"google.golang.org/grpc"

	client "github.com/SiriusDocs/backend/api_gateway/internal/app/grpc"
	httpapp "github.com/SiriusDocs/backend/api_gateway/internal/app/http"
	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/SiriusDocs/backend/api_gateway/internal/transport"
	authHandler "github.com/SiriusDocs/backend/api_gateway/internal/transport/handlers/auth"
	"github.com/SiriusDocs/backend/api_gateway/pkg/api/auth"
)

type App struct {
	HTTPServer *httpapp.App
	grpcConns  []*grpc.ClientConn
}

func New(log *slog.Logger, cfg *config.Config) *App {
	var conns []*grpc.ClientConn

	authConn, err := client.NewClient(cfg.Clients.AuthService.Address, log)
	if err != nil {
		panic(err)
	}
	conns = append(conns, authConn)

	// documentConn, err := grpcclient.NewClient(cfg.Clients.DocumentService.Address, log)
	// if err != nil { panic(err) }
	// conns = append(conns, documentConn)

	authClient := auth.NewAuthClient(authConn)
	authService := services.NewAuthService(authClient)

	// documentClient := billing.NewDocumentClient(documentConn)
	// documentService := services.NewDocumentService(documentClient)

	router := transport.NewRouter()
	apiGroup := router.Group("/")

	authHndlr := authHandler.New(log, authService)
	authHndlr.RegisterRoutes(apiGroup)

	// docHndlr := docHandler.New(log, documentService)
	// docHndlr.RegisterRoutes(apiGroup)

	httpServer := httpapp.New(log, cfg.HTTPServer.Port, router)

	return &App{
		HTTPServer: httpServer,
	}
}
