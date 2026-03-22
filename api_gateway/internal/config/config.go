package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string        `yaml:"env" env-default:"local"`
	HTTPServer HTTPServer    `yaml:"http_server"`
	Clients    ClientsConfig `yaml:"clients"`
}

type HTTPServer struct {
	Host        string        `yaml:"host" env-default:"localhost"`
	Port        int        `yaml:"port" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	AllowedOrigins []string      `yaml:"allowed_origins" env-default:"http://localhost:5173"` // CORS settings
}

type ClientsConfig struct {
	AuthService GRPCClient `yaml:"auth_service"`
	TempService GRPCClient `yaml:"temp_service"`
	// DocumentService, ChahgesService и т.д.
}

type GRPCClient struct {
	Address      string        `yaml:"address"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retries_count"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
			log.Print("No .env file found, falling back to environment variables")
		}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
