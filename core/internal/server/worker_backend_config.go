package server

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/workers"
)

const workerEngineConfigEnv = "MYCELIS_ENGINE_CONFIG_PATH"

func configuredWorkerExecutionBackend() workers.WorkerBackend {
	backend, err := loadWorkerExecutionBackend(os.Getenv(workerEngineConfigEnv), workers.EnvSecretResolver{})
	if err == nil {
		return backend
	}
	log.Printf("ERROR: worker execution backend unavailable: %v", err)
	return workers.NewUnavailableBackend(err)
}

func loadWorkerExecutionBackend(explicitPath string, secrets workers.SecretResolver) (workers.WorkerBackend, error) {
	path, required, err := resolveWorkerEngineConfigPath(explicitPath)
	if err != nil {
		return nil, err
	}
	if path == "" && !required {
		return workers.NewCentralBackend(), nil
	}
	cfg, err := workers.LoadEngineConfig(path)
	if err != nil {
		return nil, err
	}
	return workers.NewExecutionBackend(cfg, secrets)
}

func resolveWorkerEngineConfigPath(explicitPath string) (path string, required bool, err error) {
	if path = strings.TrimSpace(explicitPath); path != "" {
		if _, statErr := os.Stat(path); statErr != nil {
			return "", true, statErr
		}
		return path, true, nil
	}
	for _, candidate := range []string{"config/engine.yaml", "cognitive/config/engine.yaml", "../cognitive/config/engine.yaml"} {
		resolved, resolveErr := filepath.Abs(candidate)
		if resolveErr != nil {
			continue
		}
		if _, statErr := os.Stat(resolved); statErr == nil {
			return resolved, false, nil
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return "", false, statErr
		}
	}
	return "", false, nil
}
