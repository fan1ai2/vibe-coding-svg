package main

import (
	"log"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("Worker starting, redis=%s", cfg.RedisAddr)
}
