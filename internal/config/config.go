package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	Env            string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	BodyLimit      int
	AllowedOrigins []string
	TrustedProxies []string
	EnableLogger   bool
	RedisAddr      string
	RedisPass      string
	RedisDB        int
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Port:           ":" + port,
		Env:            env,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		BodyLimit:      4 * 1024 * 1024, // 4MB
		AllowedOrigins: []string{"*"},   // TODO: Restrict in production
		TrustedProxies: []string{},
		EnableLogger:   true,
		RedisAddr:      os.Getenv("REDIS_ADDR"),
		RedisPass:      os.Getenv("REDIS_PASS"),
		RedisDB:        getEnvInt("REDIS_DB", 0),
	}
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
