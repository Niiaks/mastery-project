package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database Database
	Server   Server
	ENV      string
}

type Database struct {
	DBName  string
	DBPass  string
	DBHost  string
	SSLMode string
	DBPort  int
	DBUser  string
}

type Server struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	return &Config{
		Database: Database{
			DBName:  GetEnv("DB_NAME", ""),
			DBPass:  GetEnv("DB_PASS", ""),
			DBHost:  GetEnv("DB_HOST", ""),
			DBPort:  GetEnvInt("DB_PORT", 0),
			SSLMode: GetEnv("SSL_MODE", "disable"),
			DBUser:  GetEnv("DB_USER", ""),
		},
		ENV: GetEnv("ENV", ""),
		Server: Server{
			Port:         GetEnv("SERVER_PORT", ""),
			ReadTimeout:  GetEnvInt("READ_TIMEOUT", 0),
			WriteTimeout: GetEnvInt("WRITE_TIMEOUT", 0),
			IdleTimeout:  GetEnvInt("IDLE_TIMEOUT", 0),
		},
	}, nil
}

func GetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return valueInt
}
