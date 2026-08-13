# Структура директорий

```
.
├── cmd
│   ├── auth_user
│   │   └── main.go
│   └── migrator
│       └── main.go
├── configs
│   └── server
│       └── local.yaml
├── internal
│   ├── app
│   │   ├── grpc
│   │   │   ├── app.go
│   │   │   └── messages.go
│   │   └── app.go
│   ├── config
│   │   └── config.go
│   ├── domain
│   │   ├── service_errors.go
│   │   └── users.go
│   ├── grpc
│   │   └── users
│   │       └── server.go
│   ├── services
│   │   ├── users
│   │   │   └── user_service.go
│   │   └── services.go
│   └── storage
│       ├── postgres
│       │   ├── postgres.go
│       │   └── user_operations.go
│       └── storage.go
├── migrations
│   ├── 00001_init_migration.sql
│   └── 00002_session_table.sql
├── pkg
│   ├── logger
│   │   └── logger.go
│   └── token-manager
│       ├── models.go
│       └── token-manager.go
├── dump_project.py
├── go.mod
├── go.sum
├── Makefile
└── README.md

20 directories, 25 files
```

# Файлы

`cmd/auth_user/main.go`
```go
package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/app"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	logger := logger.SetupLogger(cfg.Env)
	
	application := app.New(logger, cfg)

	go func(){
		application.GRPCServer.MustRun()
	}()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM
    <-stop

    // initiate graceful shutdown
    application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
    //application.Storage.Stop()
    logger.Info("Gracefully stopped")    
}
```

`cmd/migrator/main.go`
```go
package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	var dbURL, migrationsPath, migrationsTable, level string

	flag.StringVar(&dbURL, "db-url", "", "PostgreSQL connection URL")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.StringVar(&level, "level", "", "up or down the migration")
	flag.Parse()

	if dbURL == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}
	if level == "" {
		panic("it is impossible to understand whether to up or down migrations")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(fmt.Sprintf("failed to open PostgreSQL database: %v", err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	goose.SetTableName(migrationsTable)

	switch level {
	case "up":
		if err := goose.Up(db, migrationsPath); err != nil {
			panic(fmt.Sprintf("failed to up migrations: %v", err))
		}
	case "down":
		if err := goose.Down(db, migrationsPath); err != nil {
			panic(fmt.Sprintf("failed to down migrations: %v", err))
		}
	default:
		panic("it is impossible to understand whether to up or down migrations")
	}
}
```

`configs/server/local.yaml`
```yaml
auth_service:
  port: 44044
  timeout: 5s
  retries_count: 3
```

`internal/app/app.go`
```go
package app

import (
	"log/slog"

	grpcapp "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/app/grpc"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/config"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	_ "github.com/lib/pq"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(logger *slog.Logger, cfg *config.Config) *App {

	db, err := postgres.NewPostgresDB(cfg.Db.Host, cfg.Db.Port, cfg.Db.Username, cfg.Db.DBName, cfg.Db.Password, cfg.Db.SSLMode)
	if err != nil {
		panic(err)
	}
	manager, err := tokenmanager.NewManager(cfg.SigningKey)
	if err != nil {
		panic(err)
	}
	storage := storage.NewPostgresStorage(db)
	service := services.NewService(logger, storage, manager)
	grpcApp := grpcapp.New(logger, service, cfg.GRPC.Port)

	return &App{GRPCServer: grpcApp}
}
```

`internal/app/grpc/app.go`
```go
package grpcapp

import (
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
		PayloadRedactInterceptor(log), // безопасный логгер payload'ов
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
```

`internal/app/grpc/messages.go`
```go
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

```

`internal/config/config.go`
```go
package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string     `yaml:"env" env-default:"local"` // текущее окружение: local, dev, prod
	GRPC           GRPCConfig `yaml:"auth_service"`
	Db             DBConfig
	MigrationsPath string // путь до директории с миграциями
	SigningKey     string
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"` // таймаут обработки запросов
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, falling back to environment variables")
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		log.Fatal("MIGRATIONS_PATH is not set")
	}
	signingKey := os.Getenv("SIGNING_KEY")
	if signingKey == "" {
		log.Fatal("SIGNING_KEY is not set")
	}

	db := DBConfig{
		Host:     mustGetEnv("DB_HOST"),
		Port:     mustGetEnv("DB_PORT"),
		Username: mustGetEnv("DB_USERNAME"),
		Password: mustGetEnv("DB_PASSWORD"),
		DBName:   mustGetEnv("DB_NAME"),
		SSLMode:  mustGetEnv("DB_SSL"),
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	cfg.Db = db
	cfg.MigrationsPath = migrationsPath
	cfg.SigningKey = signingKey

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Critical error: environment variable %s is not set", key)
	}
	return value
}
```

`internal/domain/service_errors.go`
```go
package domain

import "errors"

var (
    ErrUserExists   = errors.New("user already exists")
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidCreds = errors.New("invalid credentials")

    ErrTokenNotFound = errors.New("refresh token not found")
    ErrInsertToken = errors.New("the new token cannot be written to the database")
    ErrInvalidToken = errors.New("invalid refresh token")
)
```

`internal/domain/users.go`
```go
package domain

import "time"

type User struct {
	Id                int64
	Username          string
	Email             string
	Creation_timestamp time.Time
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
```

`internal/grpc/users/server.go`
```go
package users

import (
	"context"
	"errors"
	"fmt"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services"
	pb "github.com/SiriusDocs/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersServer struct {
	pb.UnimplementedAuthServer
	services services.UserOperations
}

type AuthServer interface {
	Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error)
	Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error)
	GetNewTokens(ctx context.Context, in *pb.TokensRequest) (*pb.TokenResponce, error)
}

func Registered(gRPCServer *grpc.Server, Service services.UserOperations) {
	pb.RegisterAuthServer(gRPCServer, &UsersServer{services: Service})
}

func (u *UsersServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}

	id, err := u.services.CreateUser(ctx, in.Username, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.RegisterResponse{
		UserId: id,
	}, nil
}

func (u *UsersServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}
	_, tokens, err := u.services.GenerateTokens(ctx, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInsertToken) || errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		} else if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (u *UsersServer) GetNewTokens(ctx context.Context, in *pb.TokensRequest) (*pb.TokenResponce, error) {
	if in.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}
	tokens, err := u.services.RefreshToken(ctx, in.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInsertToken) || errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		} else if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		fmt.Println(err)
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return &pb.TokenResponce{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
```

`internal/services/services.go`
```go
package services

import (
	"context"
	"log/slog"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/services/users"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
)

type UserOperations interface {
	CreateUser(ctx context.Context, username string, email string, password string) (int64, error)
	GenerateTokens(ctx context.Context, email string, password string) (int64, domain.Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (domain.Tokens, error)
}

type Service struct {
	UserOperations
}

func NewService(logger *slog.Logger, store *storage.Storage, manager *tokenmanager.Manager) *Service {
	return &Service{
		UserOperations: users.NewUserOperations(logger, store.UserOperations, manager),
	}
}

```

`internal/services/users/user_service.go`
```go
package users

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log/slog"
	"os"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 87600 * time.Minute
)

type UsersService struct{
	log *slog.Logger
	store storage.UserOperations
	manager *tokenmanager.Manager
}

func NewUserOperations(logger *slog.Logger, store storage.UserOperations, manager *tokenmanager.Manager) *UsersService {
	return &UsersService{
		log: logger,
		store: store,
		manager: manager,
	}
}

func (u *UsersService) CreateUser(ctx context.Context, username string, email string, password string) (int64, error){
	id, err := u.store.CreateUser(ctx, username,email,generatePasswordHash(password))
	return id, err
}

func (u *UsersService) GenerateTokens(ctx context.Context, email string, password string) (int64, domain.Tokens, error) {
	user, err := u.store.GetUser(ctx, email, generatePasswordHash(password))
	if err != nil {
		fmt.Println(err)
		return 0,domain.Tokens{},err
	}
	return u.createSession(ctx, user)
}

func (u *UsersService) RefreshToken(ctx context.Context, refreshToken string) (domain.Tokens, error) {
	userId, err := u.store.IsTokenValid(ctx, refreshToken)
	if err != nil {
		return domain.Tokens{}, err
	}

	user := domain.User{Id: userId}
	_, tokens, err := u.createSession(ctx, user)

	return tokens, err
}

//--------------------

func (u *UsersService) createSession(ctx context.Context, user domain.User) (int64, domain.Tokens, error) {
	var (
		res domain.Tokens
		err error
	)
	res.AccessToken, err = u.manager.NewJWT(user, accessTokenTTL)
	if err != nil {
		return 0, domain.Tokens{}, err
	}
	res.RefreshToken, err = u.manager.NewRefreshToken()
	if err != nil {
		return 0, domain.Tokens{}, err
	}

	session := tokenmanager.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(refreshTokenTTL),
	}
	err = u.store.SetSession(ctx, user.Id, session)
	return user.Id, res, err
}


func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(os.Getenv("SALT"))))

}
```

`internal/storage/postgres/postgres.go`
```go
package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	usersTable = "users"
	refreshSessionsTable = "refresh_sessions"
)

func NewPostgresDB(host string, port string, username string, DBName string, password string, sslMode string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", host, port, username, DBName, password, sslMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
```

`internal/storage/postgres/user_operations.go`
```go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserOperationsPostgres struct {
	db *sqlx.DB
}

func NewUserOperationsPostgres(db *sqlx.DB) *UserOperationsPostgres {
	return &UserOperationsPostgres{
		db: db,
	}
}

func (u *UserOperationsPostgres) CreateUser(ctx context.Context, username string, email string, password string) (int64, error) {
	var id int64
	query := fmt.Sprintf("INSERT INTO %s (username, email, hashed_password) values ($1, $2, $3) RETURNING id", usersTable)

	if err := u.db.GetContext(ctx, &id, query, username, email, password); err != nil {
		// Unique Violation
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // 23505 — код unique_violation
				return 0, domain.ErrUserExists
			}
		}
		return 0, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (u *UserOperationsPostgres) GetUser(ctx context.Context, email string, password string) (domain.User, error) {
	var user domain.User
	query := fmt.Sprintf("SELECT id,username,email,creation_timestamp FROM %s WHERE email=$1 AND hashed_password=$2", usersTable)
	if err := u.db.GetContext(ctx, &user, query, email, password); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return domain.User{}, domain.ErrUserNotFound
        }
        return domain.User{}, fmt.Errorf("repository: failed to get user: %w", err)
	}
	return user, nil
}

func (u *UserOperationsPostgres) SetSession(ctx context.Context, userId int64, session tokenmanager.Session) error {
	// Используем ExecContext, так как нам не нужно возвращать данные, только проверить ошибку
	// ON CONFLICT (user_id) означает: "Если запись с таким user_id уже есть..."
	// DO UPDATE SET ... означает: "...то обнови поля refresh_token и expires_in"
	
	query := fmt.Sprintf(`
		INSERT INTO %s (user_id, refresh_token, expires_in) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			refresh_token = EXCLUDED.refresh_token, 
			expires_in = EXCLUDED.expires_in
	`, refreshSessionsTable)

	_, err := u.db.ExecContext(ctx, query, userId, session.RefreshToken, session.ExpiresAt)
	
	if err != nil {
		return fmt.Errorf("failed to set session: %w", domain.ErrInsertToken)
	}

	return nil
}

func (u *UserOperationsPostgres) IsTokenValid(ctx context.Context, refreshToken string) (int64, error) {
	var userID int64
	now := time.Now()
	query := fmt.Sprintf("SELECT r.user_id FROM %s r JOIN %s u ON u.id = r.user_id WHERE r.refresh_token=$1 AND r.expires_in>$2", refreshSessionsTable, usersTable)

	row := u.db.QueryRowContext(ctx, query, refreshToken, now)

	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrInvalidToken
		}
		return 0, fmt.Errorf("db error: %w", err)
	}
	return userID, nil
}
```

`internal/storage/storage.go`
```go
package storage

import (
	"context"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/storage/postgres"
	tokenmanager "git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/pkg/token-manager"
	"github.com/jmoiron/sqlx"
)

type UserOperations interface {
	CreateUser(ctx context.Context, username string, email string, password string) (int64, error)
	GetUser(ctx context.Context, email string, password string) (domain.User, error)
	SetSession(ctx context.Context, userId int64, session tokenmanager.Session) error
	IsTokenValid(ctx context.Context, refreshToken string) (int64, error)
}

type Storage struct {
	UserOperations
}

func NewPostgresStorage(db *sqlx.DB) *Storage {
	return &Storage{
		UserOperations: postgres.NewUserOperationsPostgres(db),
	}
}
```

`Makefile`
```makefile
APP_NAME=auth-service
MAIN_PATH=cmd/app/main.go
CONFIG_PATH=configs/server/local.yaml

.PHONY: run gen-auth swagger-gen

run:
	go run $(MAIN_PATH)

gen-auth:
	protoc -I api/proto api/proto/auth.proto \
	--go_out=pkg/api/ --go_opt=paths=source_relative \
	--go-grpc_out=pkg/api/ --go-grpc_opt=paths=source_relative



```

`migrations/00001_init_migration.sql`
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
	id serial PRIMARY KEY,
	username VARCHAR(255) NOT NULL UNIQUE,
	email VARCHAR(255) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    creation_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
```

`migrations/00002_session_table.sql`
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_sessions (
    "id" SERIAL PRIMARY KEY,
    "user_id" INT REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    "refresh_token" VARCHAR(128) NOT NULL UNIQUE,
    "expires_in" timestamp with time zone NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE refresh_sessions;
-- +goose StatementEnd
```

`pkg/logger/logger.go`
```go
package logger

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
```

`pkg/token-manager/models.go`
```go
package tokenmanager

import "time"

type Session struct {
	RefreshToken string
	ExpiresAt    time.Time
}
```

`pkg/token-manager/token-manager.go`
```go
package tokenmanager

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"git.wolkodaf2946.ru/Wolkodaf/microservices_prac/auth_service/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	NewJWT(user domain.User, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

type Manager struct {
	signingKey string
}

type tokenClaims struct {
	jwt.RegisteredClaims
	UserId   int64 `json:"user_id"`
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}
	return &Manager{signingKey: signingKey}, nil
}

//-----------------------

func (m *Manager) NewJWT(user domain.User, ttl time.Duration) (string, error) {

	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserId:   user.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(m.signingKey))
	if err != nil {
		return "", fmt.Errorf("token signature error: %w", err)
	}

	return signedToken, nil
}

//-----------------------

func (m *Manager) Parse(accessToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(m.signingKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

//-----------------------

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	_, err := r.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}
```