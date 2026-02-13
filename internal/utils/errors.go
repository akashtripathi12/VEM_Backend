package utils

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// AppError represents a custom application error
type AppError struct {
	StatusCode int
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		StatusCode: fiber.StatusBadRequest,
		Message:    message,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		StatusCode: fiber.StatusNotFound,
		Message:    message,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string) *AppError {
	return &AppError{
		StatusCode: fiber.StatusInternalServerError,
		Message:    message,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		StatusCode: fiber.StatusUnauthorized,
		Message:    message,
	}
}

// GlobalErrorHandler is the centralized error handler for Fiber
func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	// Default to 500 internal server error
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Check if it's our custom AppError
	if e, ok := err.(*AppError); ok {
		code = e.StatusCode
		message = e.Message
	}

	// Log the error
	log.Printf("Error: %v", err)

	// Send error response
	return c.Status(code).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}
