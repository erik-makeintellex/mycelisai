package mcp

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LibraryPackageTransport describes the transport definition attached to a published MCP package.
type LibraryPackageTransport struct {
	Type string `json:"type" yaml:"type"`
}

// LibraryPackage describes a registry-published package for an MCP server entry.
type LibraryPackage struct {
	RegistryType string                  `json:"registry_type" yaml:"registry_type"`
	Identifier   string                  `json:"identifier" yaml:"identifier"`
	Version      string                  `json:"version,omitempty" yaml:"version,omitempty"`
	Transport    LibraryPackageTransport `json:"transport" yaml:"transport"`
}

// LibraryEnvVar describes a declared environment variable for a curated MCP entry.
type LibraryEnvVar struct {
	Name         string `json:"name" yaml:"name"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Required     bool   `json:"required,omitempty" yaml:"required,omitempty"`
	Secret       bool   `json:"secret,omitempty" yaml:"secret,omitempty"`
	DefaultValue string `json:"default_value,omitempty" yaml:"default_value,omitempty"`
}

// LibraryEntry describes a pre-configured MCP server available for one-click install.
type LibraryEntry struct {
	Name                 string            `json:"name" yaml:"name"`
	Title                string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description          string            `json:"description" yaml:"description"`
	Version              string            `json:"version,omitempty" yaml:"version,omitempty"`
	Transport            string            `json:"transport" yaml:"transport"`
	Command              string            `json:"command" yaml:"command"`
	Args                 []string          `json:"args" yaml:"args"`
	Env                  map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	EnvironmentVariables []LibraryEnvVar   `json:"environment_variables,omitempty" yaml:"environment_variables,omitempty"`
	URL                  string            `json:"url,omitempty" yaml:"url,omitempty"`
	Packages             []LibraryPackage  `json:"packages,omitempty" yaml:"packages,omitempty"`
	Repository           string            `json:"repository,omitempty" yaml:"repository,omitempty"`
	Homepage             string            `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Tags                 []string          `json:"tags" yaml:"tags"`
	ToolSet              string            `json:"tool_set,omitempty" yaml:"tool_set,omitempty"` // suggested tool set name
	DeploymentBoundary   string            `json:"deployment_boundary,omitempty" yaml:"deployment_boundary,omitempty"`
}

// LibraryCategory groups related MCP servers.
type LibraryCategory struct {
	Name    string         `json:"name" yaml:"name"`
	Servers []LibraryEntry `json:"servers" yaml:"servers"`
}

// Library holds the full curated MCP server catalogue.
type Library struct {
	Categories []LibraryCategory `json:"categories" yaml:"categories"`
}

// LoadLibrary reads and parses the curated MCP library YAML file.
func LoadLibrary(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read library file %s: %w", path, err)
	}

	var lib Library
	if err := yaml.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("parse library file %s: %w", path, err)
	}

	return &lib, nil
}

// FindByName searches across all categories for a server with the given name.
func (lib *Library) FindByName(name string) *LibraryEntry {
	for _, cat := range lib.Categories {
		for i := range cat.Servers {
			if cat.Servers[i].Name == name {
				return &cat.Servers[i]
			}
		}
	}
	return nil
}

// ToServerConfig converts a LibraryEntry into a ServerConfig suitable for Install().
// envOverrides allows the caller to fill in required environment variables.
func (entry *LibraryEntry) ToServerConfig(envOverrides map[string]string) ServerConfig {
	env := make(map[string]string, len(entry.Env)+len(entry.EnvironmentVariables))
	for k, v := range entry.Env {
		env[k] = v
	}
	for _, spec := range entry.EnvironmentVariables {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			continue
		}
		if _, exists := env[name]; exists {
			continue
		}
		env[name] = spec.DefaultValue
	}
	// Apply overrides (user-provided values for required env vars)
	for k, v := range envOverrides {
		env[k] = v
	}

	return ServerConfig{
		Name:      entry.Name,
		Transport: entry.Transport,
		Command:   entry.Command,
		Args:      entry.Args,
		Env:       env,
		URL:       entry.URL,
	}
}

// DeclaredEnvKeys returns the canonical environment variable keys for the entry.
func (entry *LibraryEntry) DeclaredEnvKeys() []string {
	seen := make(map[string]struct{}, len(entry.Env)+len(entry.EnvironmentVariables))
	keys := make([]string, 0, len(entry.Env)+len(entry.EnvironmentVariables))
	for key := range entry.Env {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		keys = append(keys, trimmed)
	}
	for _, spec := range entry.EnvironmentVariables {
		trimmed := strings.TrimSpace(spec.Name)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		keys = append(keys, trimmed)
	}
	return keys
}

// HasRequiredSecretEnvVar reports whether the entry declares any required secret credential.
func (entry *LibraryEntry) HasRequiredSecretEnvVar() bool {
	for _, spec := range entry.EnvironmentVariables {
		if !spec.Required || !spec.Secret {
			continue
		}
		if strings.TrimSpace(spec.Name) == "" {
			continue
		}
		return true
	}
	return false
}
