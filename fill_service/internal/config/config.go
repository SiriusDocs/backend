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
	GRPC           GRPCConfig `yaml:"temp_service"`
	Db             DBConfig
	MigrationsPath string // путь до директории с миграциями
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
