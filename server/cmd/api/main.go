package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/migrate"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/router"
	_ "github.com/lib/pq"
)

// @title           SVG 转换器 API
// @version         1.0
// @description     位图转 SVG 矢量图转换服务
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}
	log.Println("已连接到 PostgreSQL")

	// 执行数据库迁移
	if err := migrate.Run(db, "migrations"); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 设置路由
	r := router.Setup(cfg, db)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 在 goroutine 中启动服务
	go func() {
		log.Printf("API 服务器启动中，端口: :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("收到信号 %v，正在优雅关闭...", sig)

	// 最多等待 30 秒让现有请求处理完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	log.Println("服务器已安全关闭")
}
