package router

import (
	"database/sql"
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	_ "github.com/fan1ai2/vibe-coding-svg/server/docs"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/handler"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/middleware"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()

	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)
	authH := handler.NewAuthHandler(cfg, authSvc)

	storage, err := service.NewStorage(cfg)
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketOriginals); err != nil {
		log.Fatalf("bucket originals: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketResults); err != nil {
		log.Fatalf("bucket results: %v", err)
	}

	convRepo := repo.NewConversionRepo(db)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	convSvc := service.NewConversionService(cfg, convRepo, storage, asynqClient)
	convH := handler.NewConversionHandler(cfg, convSvc)

	healthH := handler.NewHealthHandler(db)
	fileH := handler.NewFileHandler(storage)

	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100))

	r.GET("/health", healthH.Check)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		files := api.Group("/files")
		files.Use(middleware.JWTAuth(cfg))
		{
			files.GET("/:bucket/*key", fileH.Serve)
		}

		auth := api.Group("/auth")
		{
			auth.GET("/github/login", authH.GithubLogin)
		auth.GET("/github/callback", authH.GithubCallback)
			auth.POST("/refresh", middleware.JWTAuth(cfg), authH.Refresh)
			auth.GET("/me", middleware.JWTAuth(cfg), authH.Me)
		}

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
