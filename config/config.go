package config

import (
	"os"
)

const (
	port         = "PORT"
	postgresConn = "MARKETPLACE_POSTGRES_CONN"
)

type Config struct {
	Port         string
	PostgresConn string
}

var config *Config

func getEnvOrDefault(env string, defaultVal string) string {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	}
	return e
}

func GetConfiguration() *Config {
	if config != nil {
		return config
	}

	config := &Config{
		Port:         getEnvOrDefault(port, "8080"),
		PostgresConn: getEnvOrDefault(postgresConn, "host=localhost port=5432 user=postgres password=test-grpc dbname=chat-grpc sslmode=disable"),
	}

	return config
}
