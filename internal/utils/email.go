package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
)

// EmailBridgePayload defines the body structure for Google Apps Script
type EmailBridgePayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendEmail sends an email using Google Apps Script Web App bridge
func SendEmail(cfg *config.Config, to []string, subject string, body string) error {
	scriptURL := cfg.GoogleScriptURL
	if scriptURL == "" {
		return fmt.Errorf("GOOGLE_SCRIPT_URL missing - cannot send email")
	}

	// Google Script POST only supports one 'to' at a time in our current script
	// or we join them. Since our worker typically sends one per guest/family,
	// we'll use the first one.
	targetEmail := to[0]

	payload := EmailBridgePayload{
		To:      targetEmail,
		Subject: subject,
		Body:    body,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", scriptURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second, // Google Scripts can be slow
	}

	log.Printf("📡 [DEBUG] Sending Google Script bridge request (Target: %s)", targetEmail)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ [DEBUG] Google Script Call Failed: %v", err)
		return fmt.Errorf("failed to call Google Script: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 [DEBUG] Google Script Response Status: %d | Body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("Google Script error (Status %d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("📧 Email sent to %s via Google Script bridge", targetEmail)
	return nil
}
