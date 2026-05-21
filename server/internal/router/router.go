package router

import (
	"database/sql"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/handler"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/middleware"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.Default()

	userRepo := repo.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)
	authH := handler.NewAuthHandler(cfg, authSvc)

	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/github/login", authH.GithubLogin)
			auth.GET("/github/callback", authH.GithubCallback)
			auth.GET("/google/login", authH.GoogleLogin)
			auth.GET("/google/callback", authH.GoogleCallback)
			auth.POST("/refresh", middleware.JWTAuth(cfg), authH.Refresh)
			auth.GET("/me", middleware.JWTAuth(cfg), authH.Me)
		}
	}

	return r
}
