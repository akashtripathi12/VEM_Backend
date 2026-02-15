package store

import (
	"context"
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis(cfg *config.Config) {
	if cfg.RedisAddr == "" {
		log.Println("⚠️ Redis address not provided, skipping Redis initialization")
		return
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Println("❌ Failed to connect to Redis:", err)
	} else {
		log.Println("✅ Connected to Redis at", cfg.RedisAddr)

		// Set LFU strategy
		err = RDB.ConfigSet(ctx, "maxmemory-policy", "allkeys-lfu").Err()
		if err != nil {
			log.Println("⚠️ Failed to set Redis eviction policy to LFU:", err)
		} else {
			log.Println("✅ Redis eviction policy set to allkeys-lfu")
		}
	}
}
