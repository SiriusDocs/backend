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
	GRPC           GRPCConfig `yaml:"grpc"`
	S3             S3         `yaml:"s3"`
	MigrationsPath string     // путь до директории с миграциями
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"` // таймаут обработки запросов
}

type S3 struct {
	Endpoint       string        `yaml:"endpoint" env:"S3_ENDPOINT"`
	Region         string        `yaml:"region" env:"S3_REGION" env-default:"ru-1"`
	Bucket         string        `yaml:"bucket" env:"S3_BUCKET"`
	MaxUploadSize  int64         `yaml:"max_upload_size" env:"S3_MAX_UPLOAD_SIZE" env-default:"524288000"` // 500 MiB
	UploadURLTTL   time.Duration `yaml:"upload_url_ttl" env:"S3_UPLOAD_URL_TTL" env-default:"15m"`
	DownloadURLTTL time.Duration `yaml:"download_url_ttl" env:"S3_DOWNLOAD_URL_TTL" env-default:"15m"`
	AllowedTypes   []string      `yaml:"allowed_content_types" env:"S3_ALLOWED_CONTENT_TYPES"`

	AccessKey string
	SecretKey string
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found, falling back to environment variables")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/server/local.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	cfg.S3.AccessKey = mustGetEnv("S3_ACCESS_KEY")
	cfg.S3.SecretKey = mustGetEnv("S3_SECRET_KEY")

	if cfg.S3.Endpoint == "" {
		log.Fatal("S3 endpoint is not set (yaml s3.endpoit)")
	}
	if cfg.S3.Bucket == "" {
		log.Fatal("S3 bucket is not set (yaml s3.bucket)")
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
