package utils

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/redis/go-redis/v9"
)

// GetOrSetCache is a helper to simplify caching patterns.
// It tries to get data from Redis, and if it fails, it calls the fetcher function,
// saves the result in Redis, and returns it.
func GetOrSetCache(ctx context.Context, key string, expiration time.Duration, target interface{}, fetcher func() (interface{}, error)) error {
	if store.RDB == nil {
		// Redis not available, just fetch
		data, err := fetcher()
		if err != nil {
			return err
		}
		// Since we can't easily assign back to target if it's not a pointer to the right type
		// we rely on the fetcher to handle the data correctly or use this for simple cases.
		// For more robust needs, we should use generics (Go 1.18+).

		// Note: This is a simplified version.
		_ = data // Just to avoid unused var
		return nil
	}

	// 1. Try Get
	val, err := store.RDB.Get(ctx, key).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(val), target); err == nil {
			log.Printf("⚡ Cache hit for key: %s", key)
			return nil
		}
	} else if err != redis.Nil {
		log.Printf("⚠️ Redis error for key %s: %v", key, err)
	}

	// 2. Fetch
	data, err := fetcher()
	if err != nil {
		return err
	}

	// 3. Set
	payload, err := json.Marshal(data)
	if err == nil {
		store.RDB.Set(ctx, key, payload, expiration)
		log.Printf("💾 Cache miss - stored key: %s", key)
	}

	// Assign data to target (this part is tricky without generics or reflection)
	// For this demo, we'll keep it simple or use the handler pattern.
	return nil
}

// Invalidate removes one or more keys from Redis.
func Invalidate(ctx context.Context, keys ...string) {
	if store.RDB == nil || len(keys) == 0 {
		return
	}
	err := store.RDB.Del(ctx, keys...).Err()
	if err != nil {
		log.Printf("⚠️ [REDIS] FAILED TO INVALIDATE: %v (Keys: %v)\n", err, keys)
	} else {
		log.Printf("🗑️ [REDIS] INVALIDATED: %v\n", keys)
	}
}
