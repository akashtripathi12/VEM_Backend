package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/akashtripathi12/TBO_Backend/internal/utils"
	"github.com/hibiken/asynq"
)

// TaskHandler holds dependencies for task processing
type TaskHandler struct {
	Cfg *config.Config
}

// HandleEmailTask processes the email delivery task
func (h *TaskHandler) HandleEmailTask(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("📨 [WORKER] Processing email task for: %s", p.To)

	// Use our existing email utility
	err := utils.SendEmail(h.Cfg, []string{p.To}, p.Subject, p.Body)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
