package cognitive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// Wrapper for the robust CQA loop
func (r *Router) InferWithContract(ctx context.Context, req InferRequest) (*InferResponse, error) {
	// 1. Resolve Profile
	profile, ok := r.Config.Profiles[req.Profile]
	if !ok {
		return nil, fmt.Errorf("unknown profile: %s", req.Profile)
	}

	// 2. Loop
	var lastErr error
	currentModelID := profile.ActiveModel

	// We iterate 0..MaxRetries
	// Attempt 0 is the first try.
	for attempt := 0; attempt <= profile.MaxRetries; attempt++ {
		// 2a. Resolve Model Object
		model, err := r.getModel(currentModelID)
		if err != nil {
			return nil, err
		}

		// 2b. Setup Timeout Context for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, profile.GetTimeout())

		// 2c. Prepare Prompt (Inject correction if retrying)
		adjustedPrompt := req.Prompt
		if lastErr != nil && attempt > 0 {
			log.Printf("[INFO] Self-Correction | Profile: %s | Attempt: %d | Error: %v", req.Profile, attempt, lastErr)
			adjustedPrompt = fmt.Sprintf("%s\n\nSYSTEM: Your previous response was invalid: %s. FIX IT.", req.Prompt, lastErr.Error())
		}

		// 2d. Execute
		start := time.Now()
		resp, err := r.executeRequest(attemptCtx, model, adjustedPrompt, profile.Temperature)
		cancel() // Release context resources immediately

		elapsed := time.Since(start)

		// 2e. Handle Network/Timeout Errors
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Printf("[WARN] Cognitive Timeout | Profile: %s | Elapsed: %dms", req.Profile, elapsed.Milliseconds())
				lastErr = fmt.Errorf("timeout exceeded (%dms)", elapsed.Milliseconds())
			} else {
				log.Printf("[WARN] Provider Error | Profile: %s | %v", req.Profile, err)
				lastErr = err
			}
			// Continue to retry loop
			continue
		}

		// 2f. Validation (Output Schema)
		if err := r.validateCurrent(resp.Text, profile.OutputSchema); err != nil {
			log.Printf("[WARN] Validation Failed | Profile: %s | %v", req.Profile, err)
			lastErr = err
			continue // Retry with correction
		}

		// Success!
		return resp, nil
	}

	// 3. Fallback?
	if profile.FallbackModel != "" && profile.FallbackModel != currentModelID {
		log.Printf("[WARN] Primary Exhausted. Switching to Fallback: %s", profile.FallbackModel)
		// One last try with fallback, no retries for now (or maybe 1)
		model, err := r.getModel(profile.FallbackModel)
		if err == nil {
			// Quick execution without correction history for simplicity
			// In real sys, we might want to keep the history.
			ctxFallback, cancel := context.WithTimeout(ctx, profile.GetTimeout())
			defer cancel()
			return r.executeRequest(ctxFallback, model, req.Prompt, profile.Temperature)
		}
	}

	return nil, fmt.Errorf("cqa failure: exceeded max retries. last error: %v", lastErr)
}

func (r *Router) validateCurrent(text string, schema string) error {
	if schema == "" {
		return nil
	}

	text = strings.TrimSpace(text)

	switch schema {
	case "boolean":
		lower := strings.ToLower(text)
		if lower != "true" && lower != "false" {
			// Try parsing json {"value": true} ?
			// For now strict text
			return fmt.Errorf("expected 'true' or 'false', got: %s", text)
		}
	case "strict_json", "json_blueprint":
		if !strings.HasPrefix(text, "{") && !strings.HasPrefix(text, "[") {
			return fmt.Errorf("output does not start with JSON object/array")
		}
		var js interface{}
		if err := json.Unmarshal([]byte(text), &js); err != nil {
			return fmt.Errorf("invalid json: %w", err)
		}
	}
	return nil
}

func (r *Router) getModel(id string) (*Model, error) {
	for _, m := range r.Config.Models {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("model definition not found: %s", id)
}
