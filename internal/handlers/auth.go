package handlers

import (
	"errors"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (m *Repository) LoginAgent(c *fiber.Ctx) error {
	// TODO: Implement agent login logic
	// 1. Parse request body
	// 2. Validate credentials against m.DB.GetAgentCredentials()
	// 3. Generate JWT token
	// 4. Return token + user info

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Agent Login Endpoint",
		"token":   "placeholder-jwt-token",
	})
}

func (m *Repository) LoginGuest(c *fiber.Ctx) error {
	// TODO: Implement guest login logic
	// 1. Parse request body
	// 2. Validate guest credentials
	// 3. Generate JWT token
	// 4. Return token + user info

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Guest Login Endpoint",
		"token":   "placeholder-jwt-token",
	})
}

func (m *Repository) Logout(c *fiber.Ctx) error {
	// TODO: Implement logout logic
	// 1. Invalidate JWT/session token
	// 2. Clear cookies/headers

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "Logout successful",
	})
}

// LoginRequest defines the payload for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupRequest defines the payload for agent signup
type SignupRequest struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Phone         string `json:"phone"`
	AgencyName    string `json:"agencyName"`
	AgencyCode    string `json:"agencyCode"`
	Location      string `json:"location"`
	BusinessPhone string `json:"businessPhone"`
}

// LoginHandler authenticates a user and returns a JWT
func (r *Repository) LoginHandler(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	if err := r.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// SignupHandler registers a new agent (replaces OnboardAgent)
func (r *Repository) SignupHandler(c *fiber.Ctx) error {
	var req SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if user exists
	var existingUser models.User
	if err := r.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	newUser := models.User{
		ID:           uuid.New(), // Explicitly generate UUID to avoid DB default issues
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "agent",
		Name:         req.Name,
		Phone:        req.Phone,
	}

	err = r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		newProfile := models.AgentProfile{
			UserID:        newUser.ID,
			AgencyName:    req.AgencyName,
			AgencyCode:    req.AgencyCode,
			Location:      req.Location,
			BusinessPhone: req.BusinessPhone,
		}

		if err := tx.Create(&newProfile).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Log the actual error for debugging
		println("Error creating user:", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user: " + err.Error()})
	}

	// Generate Token for immediate login
	token, _ := utils.GenerateToken(newUser.ID, newUser.Email, newUser.Role)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User allocated successfully",
		"token":   token,
		"user":    newUser,
	})
}

// GetMe returns the current authenticated user
func (r *Repository) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var user models.User
	if err := r.DB.Preload("AgentProfile").First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}
