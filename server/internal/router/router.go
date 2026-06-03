package router

import (
	"database/sql"
	"log"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/ai"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/handler"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/middleware"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/neo4j"
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

	// 启动清理过期验证码的后台协程
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			userRepo.CleanupExpiredCodes()
		}
	}()

	emailSvc := service.NewEmailService(cfg)
	authSvc := service.NewAuthService(cfg, userRepo, emailSvc)
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

	// --- Neo4j ---
	if err := neo4j.Init(cfg.Neo4jURI, "neo4j", cfg.Neo4jPassword); err != nil {
		log.Fatalf("Neo4j 连接失败: %v", err)
	}

	// --- 图标库模块 ---
	iconRepo := repo.NewIconRepo(db)
	tagRepo := repo.NewTagRepo(db)
	graphSync := neo4j.NewGraphSyncService()
	iconSvc := service.NewIconService(iconRepo, tagRepo, graphSync)
	iconH := handler.NewIconHandler(iconSvc)
	tagH := handler.NewTagHandler(tagRepo)

	// --- AI 图标生成模块 ---
	promptBuilder := ai.NewPromptBuilder(db)
	aiProvider := ai.NewOpenAIClient(cfg.AiBaseURL, cfg.AiApiKey, cfg.AiModel, promptBuilder)
	aiQuota := ai.NewQuotaService(cfg.RedisAddr)
	aiSvc := service.NewAiService(aiProvider, aiQuota)
	aiH := handler.NewAiHandler(aiSvc)

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
			// 访客
			auth.POST("/guest", authH.GuestLogin)

			// 邮箱
			auth.POST("/email/send-code", authH.EmailSendCode)
			auth.POST("/email/verify", authH.EmailVerify)

			// GitHub（已有）
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

		// 编辑器保存的 SVG（全部需要认证）
		savedSvgRepo := repo.NewSavedSvgRepo(db)
		savedSvgH := handler.NewSavedSvgHandler(savedSvgRepo)
		svgs := api.Group("/svgs")
		svgs.Use(middleware.JWTAuth(cfg))
		{
			svgs.POST("", savedSvgH.Save)
			svgs.GET("", savedSvgH.List)
			svgs.GET("/:id", savedSvgH.Get)
			svgs.GET("/:id/download", savedSvgH.Download)
			svgs.DELETE("/:id", savedSvgH.Delete)
		}

		// 图标库 — 搜索和详情公开，其余需认证
		icons := api.Group("/icons")
		{
			icons.Use(middleware.OptionalJWTAuth(cfg))
			icons.GET("/search", iconH.Search)
			icons.GET("/:id", iconH.Get)
			icons.GET("/:id/recommend", iconH.Recommend)
			icons.GET("", iconH.List)

			icons.Use(middleware.JWTAuth(cfg))
			icons.POST("", iconH.Create)
			icons.POST("/batch", iconH.BatchCreate)
			icons.DELETE("/:id", iconH.Delete)
		}

		// AI 图标生成（全部需要认证）
		aig := api.Group("/ai")
		aig.Use(middleware.JWTAuth(cfg))
		{
			aig.POST("/generate", aiH.Generate)
			aig.GET("/quota", aiH.Quota)
		}

		// 标签 — 公开
		api.GET("/tags", tagH.List)
	}

	return r
}
