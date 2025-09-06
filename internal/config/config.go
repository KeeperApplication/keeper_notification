package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	RabbitMQURL            string
	FCMCredentialsFilePath string
}

func Load() *Config {
	if os.Getenv("GO_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: Could not find .env file, using system environment variables.")
		}
	}

	return &Config{
		Port:                   getEnv("PORT", "4002"),
		RabbitMQURL:            getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		FCMCredentialsFilePath: getEnv("FCM_CREDENTIALS_FILE_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
