package config

import (
	"os"
	"strconv"
)

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
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
	SMTPFrom       string
}

func Load() *Config {
	return &Config{
		Port:           envOr("PORT", "8080"),
		DatabaseURL:    require("DATABASE_URL"),
		RedisAddr:      require("REDIS_ADDR"),
		MinioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: require("MINIO_ACCESS_KEY"),
		MinioSecretKey: require("MINIO_SECRET_KEY"),
		MinioBucket:    envOr("MINIO_BUCKET", "svgconverter"),
		JWTSecret:      require("JWT_SECRET"),
		GithubClientID: require("GITHUB_CLIENT_ID"),
		GithubSecret:   require("GITHUB_CLIENT_SECRET"),
		MaxFileSize:    intEnvOr("MAX_FILE_SIZE", 10<<20),
		FrontendURL:    require("FRONTEND_URL"),
		SMTPHost:       os.Getenv("SMTP_HOST"),
		SMTPPort:       int(intEnvOr("SMTP_PORT", 587)),
		SMTPUser:       os.Getenv("SMTP_USER"),
		SMTPPassword:   os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:       os.Getenv("SMTP_FROM"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func require(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env var " + key + " is not set")
	}
	return v
}

func intEnvOr(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}
