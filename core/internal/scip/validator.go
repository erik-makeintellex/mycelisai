package scip

import (
	"fmt"
	"os"

	"github.com/mycelis/core/pkg/scip"
	"gopkg.in/yaml.v3"
)

// ContractConfig matches contracts.yaml
type ContractConfig struct {
	Contracts []Contract `yaml:"contracts"`
}

type Contract struct {
	ID           string   `yaml:"id"`
	AllowedTypes []string `yaml:"allowed_types"`
	MaxBytes     int      `yaml:"max_bytes"`
	RequiredMeta []string `yaml:"required_meta"`
}

type Validator struct {
	contracts map[string]Contract
}

func NewValidator(configPath string) (*Validator, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contracts: %w", err)
	}

	var config ContractConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse contracts: %w", err)
	}

	m := make(map[string]Contract)
	for _, c := range config.Contracts {
		m[c.ID] = c
	}

	return &Validator{contracts: m}, nil
}

// Validate checks if the envelope adheres to the contract associated with the intent.
// For MVP, we assume Intent == ContractID (e.g. intent="std.chat").
// In real life, we might map multiple intents to one contract.
func (v *Validator) Validate(env *scip.SignalEnvelope) error {
	// 1. Basic Integrity
	if env.TraceId == "" {
		return fmt.Errorf("missing trace_id")
	}

	// 2. Find Contract
	// Using Intent as Contract ID for now
	contract, ok := v.contracts[env.Intent]
	if !ok {
		// If no contract found for intent, strict deny? or default allow?
		// SCIP Philosophy: Strict Deny.
		return fmt.Errorf("no contract defined for intent: %s", env.Intent)
	}

	// 3. Size Check
	if len(env.Payload) > contract.MaxBytes {
		return fmt.Errorf("payload size %d exceeds limit %d for %s", len(env.Payload), contract.MaxBytes, contract.ID)
	}

	// 4. Type Check
	// Convert Enum to String to check against yaml
	// scip.DataType_name is map[int32]string
	typeName := scip.DataType_name[int32(env.DataType)]
	allowed := false
	for _, t := range contract.AllowedTypes {
		if t == typeName {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("datatype %s not allowed for %s (allowed: %v)", typeName, contract.ID, contract.AllowedTypes)
	}

	return nil
}
