package config

import (
	"fmt"
	"os"
)

// Config defines the config for the application.
type Config struct {
	ServerPort string
	JWTSecret  string
}

func getEnv(key string, defaultValue string) string {
	if value, err := os.LookupEnv(key); err {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, err := os.LookupEnv(key); err {
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err != nil {
			return defaultValue
		}
		return intValue
	}
	return defaultValue
}

func LoadConfig() Config {
	return Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"),
	}
}
