package middleware

import (
	"strings"

	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT tokens and sets user context
// This is a placeholder implementation - replace with actual JWT validation
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Missing authorization header")
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.UnauthorizedResponse(c, "Invalid authorization header format")
		}

		token := parts[1]

		// TODO: Validate JWT token
		// For now, just check if token is not empty
		if token == "" {
			return utils.UnauthorizedResponse(c, "Invalid token")
		}

		// TODO: Decode token and extract user info
		// Set user in context for handlers to access
		// c.Locals("user", user)

		return c.Next()
	}
}
