package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	Env             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	BodyLimit       int
	AllowedOrigins  []string
	TrustedProxies  []string
	EnableLogger    bool
	FrontendURL     string
	RedisAddr       string
	RedisPass       string
	RedisDB         int
	GoogleScriptURL string
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

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// --- 1. Parse Allowed Origins (For Secure CORS) ---
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string

	if allowedOriginsStr == "" {
		// Safe fallback for local development
		allowedOrigins = []string{"http://localhost:3000"}
	} else {
		// Safely split multiple domains (e.g., "http://localhost:3000, https://myapp.vercel.app")
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			allowedOrigins = append(allowedOrigins, strings.TrimSpace(origin))
		}
	}

	// --- 2. Parse Redis Connection (For Render / Cloud) ---
	// Cloud providers inject REDIS_URL. We use this instead of REDIS_ADDR.
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Fallback for local terminal testing
		redisURL = "redis://127.0.0.1:6379"
	}

	return &Config{
		Port:            ":" + port,
		Env:             env,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		BodyLimit:       4 * 1024 * 1024, // 4MB
		AllowedOrigins:  allowedOrigins,
		TrustedProxies:  []string{},
		EnableLogger:    true,
		FrontendURL:     frontendURL,
		RedisAddr:       redisURL, // Now correctly holds the URI (e.g., redis://red-xxxxx:6379)
		RedisPass:       os.Getenv("REDIS_PASS"),
		RedisDB:         getEnvInt("REDIS_DB", 0),
		GoogleScriptURL: os.Getenv("GOOGLE_SCRIPT_URL"),
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
