package server

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycelis/core/internal/outputvalidation"
)

func configuredOutputValidator() outputvalidation.Validator {
	script := strings.TrimSpace(os.Getenv("MYCELIS_OUTPUT_VALIDATOR_SCRIPT"))
	if script == "" {
		for _, candidate := range []string{
			"interface/scripts/validate-generated-output.mjs",
			"../interface/scripts/validate-generated-output.mjs",
			"/core/output-validation/validate-generated-output.mjs",
		} {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				script = candidate
				break
			}
		}
	}
	if script == "" {
		log.Printf("[output-validation] browser adapter unavailable: validator script not found")
		return nil
	}
	workingDir := strings.TrimSpace(os.Getenv("MYCELIS_OUTPUT_VALIDATOR_WORKDIR"))
	if workingDir == "" {
		workingDir = filepath.Dir(script)
	}
	validator, err := outputvalidation.NewPlaywrightValidator(outputvalidation.PlaywrightConfig{
		NodeBinary: strings.TrimSpace(os.Getenv("MYCELIS_OUTPUT_VALIDATOR_NODE")),
		ScriptPath: script,
		WorkingDir: workingDir,
		Timeout:    45 * time.Second,
	})
	if err != nil {
		log.Printf("[output-validation] browser adapter unavailable: %v", err)
		return nil
	}
	return validator
}
