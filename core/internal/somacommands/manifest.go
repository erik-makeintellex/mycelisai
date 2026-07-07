package somacommands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultConfigEnv = "MYCELIS_SOMA_COMMANDS_CONFIG"

// Registry is the config-backed source of user-facing metadata for built-in
// Soma commands. Go still owns handler execution.
type Registry struct {
	Version  string    `yaml:"version"`
	Commands []Command `yaml:"commands"`
}

type Command struct {
	ID                string         `yaml:"id"`
	Handler           string         `yaml:"handler"`
	Title             string         `yaml:"title"`
	Summary           string         `yaml:"summary"`
	UserQuote         string         `yaml:"user_quote"`
	Category          string         `yaml:"category"`
	CapabilityID      string         `yaml:"capability_id"`
	InputSchemaRef    string         `yaml:"input_schema_ref"`
	OutputSchemaRef   string         `yaml:"output_schema_ref"`
	Scope             Scope          `yaml:"scope"`
	Governance        Governance     `yaml:"governance"`
	Delivery          Delivery       `yaml:"delivery"`
	ContextAssertions []string       `yaml:"context_assertions"`
	UIVisibility      UIVisibility   `yaml:"ui_visibility"`
	Availability      Availability   `yaml:"availability"`
	Metadata          map[string]any `yaml:"metadata"`
}

type Scope struct {
	Default string   `yaml:"default"`
	Roles   []string `yaml:"roles"`
	Hosts   []string `yaml:"hosts"`
	Groups  []string `yaml:"groups"`
}

type Governance struct {
	RiskClass        string `yaml:"risk_class"`
	ApprovalPosture  string `yaml:"approval_posture"`
	AuditRequired    bool   `yaml:"audit_required"`
	PrivateDataTouch bool   `yaml:"private_data_touch"`
}

type Delivery struct {
	OutputKinds     []string `yaml:"output_kinds"`
	ProofRequired   bool     `yaml:"proof_required"`
	RecoveryPosture string   `yaml:"recovery_posture"`
}

type UIVisibility struct {
	DefaultSurface string `yaml:"default_surface"`
	InspectOnly    bool   `yaml:"inspect_only"`
	OperatorLabel  string `yaml:"operator_label"`
}

type Availability struct {
	Mode     string   `yaml:"mode"`
	Requires []string `yaml:"requires"`
}

func LoadDefault() (Registry, error) {
	return LoadFile(DefaultPath())
}

func LoadFile(path string) (Registry, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return Registry{}, err
	}
	if stat.IsDir() {
		return LoadDir(path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, err
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return Registry{}, fmt.Errorf("decode soma command manifests: %w", err)
	}
	if err := reg.Validate(); err != nil {
		return Registry{}, err
	}
	return reg, nil
}

func LoadDir(path string) (Registry, error) {
	matches, err := filepath.Glob(filepath.Join(path, "*.yaml"))
	if err != nil {
		return Registry{}, err
	}
	sort.Strings(matches)
	merged := Registry{Version: "soma_commands.v1"}
	for _, match := range matches {
		reg, err := loadSingleFile(match)
		if err != nil {
			return Registry{}, err
		}
		if reg.Version != "" {
			merged.Version = reg.Version
		}
		merged.Commands = append(merged.Commands, reg.Commands...)
	}
	if err := merged.Validate(); err != nil {
		return Registry{}, err
	}
	return merged, nil
}

func loadSingleFile(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, err
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return Registry{}, fmt.Errorf("decode soma command manifests %s: %w", path, err)
	}
	if strings.TrimSpace(reg.Version) == "" {
		reg.Version = "soma_commands.v1"
	}
	if err := reg.Validate(); err != nil {
		return Registry{}, err
	}
	return reg, nil
}

func DefaultPath() string {
	if override := strings.TrimSpace(os.Getenv(defaultConfigEnv)); override != "" {
		return override
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("core", "config", "soma-commands")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "config", "soma-commands"))
}

func (r Registry) ByHandler() map[string]Command {
	out := make(map[string]Command, len(r.Commands))
	for _, command := range r.Commands {
		out[command.Handler] = command
	}
	return out
}

func (r Registry) Validate() error {
	if strings.TrimSpace(r.Version) == "" {
		return fmt.Errorf("soma command manifests: version is required")
	}
	seenIDs := map[string]struct{}{}
	seenHandlers := map[string]struct{}{}
	for i, command := range r.Commands {
		if strings.TrimSpace(command.ID) == "" {
			return fmt.Errorf("soma command manifests: command %d missing id", i)
		}
		if strings.TrimSpace(command.Handler) == "" {
			return fmt.Errorf("soma command manifests: command %s missing handler", command.ID)
		}
		if strings.TrimSpace(command.Title) == "" || strings.TrimSpace(command.Summary) == "" {
			return fmt.Errorf("soma command manifests: command %s missing title or summary", command.ID)
		}
		if _, exists := seenIDs[command.ID]; exists {
			return fmt.Errorf("soma command manifests: duplicate id %s", command.ID)
		}
		if _, exists := seenHandlers[command.Handler]; exists {
			return fmt.Errorf("soma command manifests: duplicate handler %s", command.Handler)
		}
		seenIDs[command.ID] = struct{}{}
		seenHandlers[command.Handler] = struct{}{}
	}
	return nil
}

func (r Registry) ValidateHandlers(handlers []string) error {
	configured := r.ByHandler()
	missing := []string{}
	for _, handler := range handlers {
		if _, ok := configured[handler]; !ok {
			missing = append(missing, handler)
		}
	}
	extra := []string{}
	known := map[string]struct{}{}
	for _, handler := range handlers {
		known[handler] = struct{}{}
	}
	for handler := range configured {
		if _, ok := known[handler]; !ok {
			extra = append(extra, handler)
		}
	}
	if len(missing) == 0 && len(extra) == 0 {
		return nil
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return fmt.Errorf("soma command manifests do not match handlers: missing=%v extra=%v", missing, extra)
}
