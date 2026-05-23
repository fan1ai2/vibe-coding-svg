package main

import (
	"database/sql"
	"log"

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

	// 设置路由并启动服务
	r := router.Setup(cfg, db)
	log.Printf("API 服务器启动中，端口: :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
