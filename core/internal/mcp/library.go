package mcp

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LibraryEntry describes a pre-configured MCP server available for one-click install.
type LibraryEntry struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Transport   string            `json:"transport" yaml:"transport"`
	Command     string            `json:"command" yaml:"command"`
	Args        []string          `json:"args" yaml:"args"`
	Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Tags        []string          `json:"tags" yaml:"tags"`
	ToolSet     string            `json:"tool_set,omitempty" yaml:"tool_set,omitempty"` // suggested tool set name
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
	env := make(map[string]string, len(entry.Env))
	for k, v := range entry.Env {
		env[k] = v
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
	}
}
