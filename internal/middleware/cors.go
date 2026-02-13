package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SetupCORS configures CORS middleware with allowed origins from config
func SetupCORS(allowedOrigins []string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     joinOrigins(allowedOrigins),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400, // 24 hours
	})
}

// joinOrigins converts slice of origins to comma-separated string
func joinOrigins(origins []string) string {
	if len(origins) == 0 {
		return "*"
	}
	result := ""
	for i, origin := range origins {
		if i > 0 {
			result += ","
		}
		result += origin
	}
	return result
}
