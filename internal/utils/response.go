package utils

import "github.com/gofiber/fiber/v2"

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse sends a successful response with data
func SuccessResponse(c *fiber.Ctx, statusCode int, data interface{}) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusBadRequest, message)
}

// NotFoundResponse sends a 404 not found response
func NotFoundResponse(c *fiber.Ctx, resource string) error {
	return ErrorResponse(c, fiber.StatusNotFound, resource+" not found")
}

// InternalErrorResponse sends a 500 internal server error response
func InternalErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, message)
}

// UnauthorizedResponse sends a 401 unauthorized response
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Unauthorized"
	}
	return ErrorResponse(c, fiber.StatusUnauthorized, message)
}
