package main

import (
	"fmt"
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
	"github.com/akashtripathi12/TBO_Backend/internal/store"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	store.InitDB()
	db := store.DB
	db.Logger = logger.Default.LogMode(logger.Silent)

	tboEmail := "tbo@test.com"
	tboPass := "Test@123"
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(tboPass), bcrypt.DefaultCost)

	tboAgent := models.User{
		Email:        tboEmail,
		PasswordHash: string(hashedPass),
		Role:         "tbo_agent",
		Name:         "TBO Manager",
		Phone:        "5555555555",
	}

	var existing models.User
	if err := db.Where("email = ?", tboEmail).First(&existing).Error; err != nil {
		tboAgent.ID = uuid.New()
		if err := db.Create(&tboAgent).Error; err != nil {
			log.Fatalf("❌ Failed to create TBO Agent: %v", err)
		}
		fmt.Println("✅ TBO Agent created successfully!")
	} else {
		tboAgent = existing
		db.Model(&tboAgent).Updates(map[string]interface{}{
			"password_hash": string(hashedPass),
			"role":          "tbo_agent",
			"name":          "TBO Manager",
		})
		fmt.Println("✅ TBO Agent already exists — password & role updated!")
	}

	fmt.Println("\n=== TBO AGENT CREDENTIALS ===")
	fmt.Printf("Email:    %s\n", tboEmail)
	fmt.Printf("Password: %s\n", tboPass)
	fmt.Printf("Role:     tbo_agent\n")
	fmt.Printf("User ID:  %s\n", tboAgent.ID)
	fmt.Println("=============================")
}
