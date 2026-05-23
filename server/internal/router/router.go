package router

import (
	"database/sql"
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/handler"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/middleware"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// Setup 初始化路由、依赖注入并返回 Gin 引擎
func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()

	// --- 认证模块 ---
	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)
	authH := handler.NewAuthHandler(cfg, authSvc)

	// --- 对象存储 ---
	storage, err := service.NewStorage(cfg)
	if err != nil {
		log.Fatalf("对象存储初始化失败: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketOriginals); err != nil {
		log.Fatalf("创建原始文件存储桶失败: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketResults); err != nil {
		log.Fatalf("创建结果文件存储桶失败: %v", err)
	}

	// --- 转换模块 ---
	convRepo := repo.NewConversionRepo(db)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	convSvc := service.NewConversionService(cfg, convRepo, storage, asynqClient)
	convH := handler.NewConversionHandler(cfg, convSvc)

	healthH := handler.NewHealthHandler(db, cfg.RedisAddr, storage.Client())
	fileH := handler.NewFileHandler(storage)

	// 全局中间件
	r.Use(middleware.CORS(cfg.FrontendURL))
	r.Use(middleware.RequestLogging())
	r.Use(middleware.RateLimit(cfg.RedisAddr, 100))

	// Swagger 文档
	r.Static("/docs", "./docs")

	// 健康检查（无需认证）
	r.GET("/health", healthH.Check)

	api := r.Group("/api/v1")
	{
		// 文件服务（公开访问，URL 中的 UUID key 不可猜测）
		files := api.Group("/files")
		{
			files.GET("/:bucket/*key", fileH.Serve)
		}

		// 认证接口（部分公开）
		auth := api.Group("/auth")
		{
			auth.GET("/github/login", authH.GithubLogin)
			auth.GET("/github/callback", authH.GithubCallback)
			auth.POST("/refresh", middleware.JWTAuth(cfg), authH.Refresh)
			auth.GET("/me", middleware.JWTAuth(cfg), authH.Me)
		}

		// 转换接口（全部需要认证）
		conversions := api.Group("/conversions")
		conversions.Use(middleware.JWTAuth(cfg))
		{
			conversions.POST("", convH.Upload)
			conversions.GET("", convH.List)
			conversions.GET("/:id", convH.Status)
			conversions.GET("/:id/download", convH.Download)
		}
	}

	return r
}
