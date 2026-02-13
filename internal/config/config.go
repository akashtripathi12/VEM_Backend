package config

import "time"

type Config struct {
	Port           string
	Env            string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	BodyLimit      int
	AllowedOrigins []string
	TrustedProxies []string
	EnableLogger   bool
}

func Load() *Config {
	return &Config{
		Port:           ":8080",
		Env:            "development",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		BodyLimit:      4 * 1024 * 1024, // 4MB
		AllowedOrigins: []string{"*"},   // TODO: Restrict in production
		TrustedProxies: []string{},
		EnableLogger:   true,
	}
}
