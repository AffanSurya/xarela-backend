package config

import "os"

const (
	defaultPort       = "8080"
	defaultLogLevel   = "info"
	defaultDatabaseDSN = ""
)

type Config struct {
	Port        string
	LogLevel    string
	DatabaseDSN string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", defaultPort),
		LogLevel:    getEnv("LOG_LEVEL", defaultLogLevel),
		DatabaseDSN: getEnv("DATABASE_DSN", defaultDatabaseDSN),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
