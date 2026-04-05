package main

import (
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	envPath := flag.String("env", ".env", "Path to .env file")
	username := flag.String("u", "admin", "Admin username")
	email := flag.String("e", "admin@admin.com", "Admin email")
	password := flag.String("p", "adminpassword", "Admin password")
	flag.Parse()

	// загрузка .env файла
	if err := godotenv.Load(*envPath); err != nil {
		log.Printf("Warning: Could not load .env file from '%s'. Falling back to system environment variables.\n", *envPath)
	}

	// считывание переменных
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL")
	salt := os.Getenv("SALT")

	if salt == "" || host == "" {
		log.Fatal("Error: Required environment variables (SALT, DB_HOST, etc.) are missing. Check your .env file.")
	}

	// фрмирование строки подключения
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, dbname, sslmode)

	// подключение к БД
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// хэширование пароля
	hashedPassword := generatePasswordHash(*password, salt)

	// создание или обновление пользователя
	query := `
		INSERT INTO users (username, email, hashed_password, user_role) 
		VALUES ($1, $2, $3, 'ADMIN')
		ON CONFLICT (email) 
		DO UPDATE SET 
			user_role = 'ADMIN',
			hashed_password = EXCLUDED.hashed_password,
			username = EXCLUDED.username;
	`

	_, err = db.ExecContext(context.Background(), query, *username, *email, hashedPassword)
	if err != nil {
		log.Fatalf("Failed to create/update admin user: %v", err)
	}

	fmt.Println("=======================================")
	fmt.Println("✅ Admin successfully created/updated!")
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Email:    %s\n", *email)
	fmt.Printf("Password: %s\n", *password)
	fmt.Println("=======================================")
}

func generatePasswordHash(password, salt string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}