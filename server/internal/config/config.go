package config

import "os"

type Config struct {
	Port           string
	DatabaseURL    string
	RedisAddr      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	JWTSecret      string
	GithubClientID string
	GithubSecret   string
	MaxFileSize    int64
	FrontendURL    string
}

func Load() *Config {
	return &Config{
		Port:           envOr("PORT", "8080"),
		DatabaseURL:    envOr("DATABASE_URL", "postgres://svguser:svgpass@localhost:5432/svgconverter?sslmode=disable"),
		RedisAddr:      envOr("REDIS_ADDR", "localhost:6379"),
		MinioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    envOr("MINIO_BUCKET", "svgconverter"),
		JWTSecret:      envOr("JWT_SECRET", "dev-secret-change-in-prod"),
		GithubClientID: os.Getenv("GITHUB_CLIENT_ID"),
		GithubSecret:   os.Getenv("GITHUB_CLIENT_SECRET"),
		MaxFileSize:    10 << 20,
		FrontendURL:    envOr("FRONTEND_URL", "http://localhost:8080"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
