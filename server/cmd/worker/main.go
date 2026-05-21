package main

import (
	"database/sql"
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/worker"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("connected to postgres")

	storage, err := service.NewStorage(cfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketOriginals); err != nil {
		log.Fatalf("bucket originals: %v", err)
	}
	if err := storage.EnsureBucket(service.BucketResults); err != nil {
		log.Fatalf("bucket results: %v", err)
	}

	convRepo := repo.NewConversionRepo(db)
	convWorker := worker.NewConversionWorker(cfg, convRepo, storage)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{Concurrency: 4},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("conversion:process", convWorker.HandleProcessTask)

	log.Printf("Worker starting, redis=%s", cfg.RedisAddr)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
