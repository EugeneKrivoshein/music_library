package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Структура конфигурации
type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPass        string
	DBName        string
	ServerAddress string
	APIURL        string
}

// Функция загрузки конфигурации
func LoadConfig(envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Не удалось загрузить .env: %v. Используются переменные окружения.", err)
	}

	return &Config{
		DBUser:        os.Getenv("DB_USER"),
		DBPass:        os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
		APIURL:        os.Getenv("API_URL"),
	}, nil
}
