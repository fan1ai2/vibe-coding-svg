package config

import "os"

// Config 应用配置结构体，所有配置项均通过环境变量注入
type Config struct {
	Port           string // 服务监听端口
	DatabaseURL    string // PostgreSQL 连接字符串
	RedisAddr      string // Redis 地址（用于 asynq 任务队列）
	MinioEndpoint  string // MinIO 对象存储地址
	MinioAccessKey string // MinIO 访问密钥
	MinioSecretKey string // MinIO 秘密密钥
	MinioBucket    string // MinIO 存储桶名称
	JWTSecret      string // JWT 签名密钥
	GithubClientID string // GitHub OAuth 客户端 ID
	GithubSecret   string // GitHub OAuth 客户端密钥
	MaxFileSize    int64  // 上传文件最大大小（字节）
	FrontendURL    string // 前端地址（用于 OAuth 回调重定向）
}

// Load 从环境变量加载配置，未设置时使用默认值
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
		MaxFileSize:    10 << 20, // 10MB
		FrontendURL:    envOr("FRONTEND_URL", "http://localhost:8080"),
	}
}

// envOr 获取环境变量，如果未设置则返回默认值
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
