package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ActionCLIConfig struct {
	APIBaseURL     string            `yaml:"api_base_url"`
	APIKey         string            `yaml:"api_key"`
	TimeoutSeconds int               `yaml:"timeout_seconds"`
	Headers        map[string]string `yaml:"headers"`
}

func defaultActionCLIConfig() ActionCLIConfig {
	return ActionCLIConfig{
		APIBaseURL:     "http://localhost:8081",
		TimeoutSeconds: 20,
		Headers:        map[string]string{},
	}
}

func discoverActionConfigPaths() []string {
	return discoverActionConfigPathsWith(os.Getenv, os.Getwd, mustUserHomeDir())
}

func discoverActionConfigPathsWith(getenv func(string) string, getwd func() (string, error), userHome string) []string {
	paths := []string{
		"/etc/mycelis/config.yaml",
		"/etc/mycelis/config.yml",
	}

	if cwd, err := getwd(); err == nil && cwd != "" {
		paths = append(paths,
			filepath.Join(cwd, "mycelis.yaml"),
			filepath.Join(cwd, "mycelis.yml"),
			filepath.Join(cwd, "config", "mycelis.yaml"),
			filepath.Join(cwd, "config", "mycelis.yml"),
		)
	}

	xdgHome := getenv("XDG_CONFIG_HOME")
	if xdgHome != "" {
		paths = append(paths,
			filepath.Join(xdgHome, "mycelis", "config.yaml"),
			filepath.Join(xdgHome, "mycelis", "config.yml"),
		)
	}

	if userHome != "" {
		paths = append(paths,
			filepath.Join(userHome, ".config", "mycelis", "config.yaml"),
			filepath.Join(userHome, ".config", "mycelis", "config.yml"),
			filepath.Join(userHome, ".mycelis", "config.yaml"),
			filepath.Join(userHome, ".mycelis", "config.yml"),
		)
	}

	// Explicit override path is highest precedence if set.
	if override := getenv("MYCELIS_CONFIG"); override != "" {
		paths = append(paths, override)
	}

	return paths
}

func mustUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func mergeActionCLIConfig(dst *ActionCLIConfig, src ActionCLIConfig) {
	if src.APIBaseURL != "" {
		dst.APIBaseURL = src.APIBaseURL
	}
	if src.APIKey != "" {
		dst.APIKey = src.APIKey
	}
	if src.TimeoutSeconds > 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
	if len(src.Headers) > 0 {
		if dst.Headers == nil {
			dst.Headers = map[string]string{}
		}
		for k, v := range src.Headers {
			dst.Headers[k] = v
		}
	}
}

func loadActionCLIConfigFromPaths(paths []string, base ActionCLIConfig) (ActionCLIConfig, []string, error) {
	cfg := base
	loaded := []string{}
	for _, p := range paths {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var fileCfg ActionCLIConfig
		if err := yaml.Unmarshal(data, &fileCfg); err != nil {
			return cfg, loaded, fmt.Errorf("parse action config %s: %w", p, err)
		}
		mergeActionCLIConfig(&cfg, fileCfg)
		loaded = append(loaded, p)
	}
	return cfg, loaded, nil
}

func loadActionCLIConfig() (ActionCLIConfig, []string, error) {
	base := defaultActionCLIConfig()
	return loadActionCLIConfigFromPaths(discoverActionConfigPaths(), base)
}

func loadActionRuntimeConfig() (ActionCLIConfig, error) {
	cfg, _, err := loadActionCLIConfig()
	if err != nil {
		return cfg, err
	}

	if envBase := os.Getenv("MYCELIS_API_URL"); envBase != "" {
		cfg.APIBaseURL = envBase
	}
	if envKey := os.Getenv("MYCELIS_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}
	if envTimeout := os.Getenv("MYCELIS_API_TIMEOUT_SEC"); envTimeout != "" {
		if n, parseErr := strconv.Atoi(envTimeout); parseErr == nil && n > 0 {
			cfg.TimeoutSeconds = n
		}
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 20
	}
	return cfg, nil
}

func resolveActionURL(baseURL, target string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("empty target")
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		if _, err := url.ParseRequestURI(target); err != nil {
			return "", fmt.Errorf("invalid target URL %q: %w", target, err)
		}
		return target, nil
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasPrefix(target, "/") {
		target = "/" + target
	}
	full := baseURL + target
	if _, err := url.ParseRequestURI(full); err != nil {
		return "", fmt.Errorf("invalid resolved URL %q: %w", full, err)
	}
	return full, nil
}

func executeActionRequest(cfg ActionCLIConfig, method, target, body string) (int, string, error) {
	reqURL, err := resolveActionURL(cfg.APIBaseURL, target)
	if err != nil {
		return 0, "", err
	}

	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(strings.ToUpper(strings.TrimSpace(method)), reqURL, reader)
	if err != nil {
		return 0, "", fmt.Errorf("build request: %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	if cfg.APIKey != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody), nil
}

func printActionResponse(statusCode int, body string) {
	fmt.Printf("HTTP %d\n", statusCode)
	if strings.TrimSpace(body) != "" {
		fmt.Println(body)
	}
}

func parseActionShellLine(line string) (method, target, body string, handled bool, err error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", "", true, nil
	}

	if line == "status" {
		return http.MethodGet, "/api/v1/services/status", "", true, nil
	}

	if strings.HasPrefix(line, "/") {
		return http.MethodGet, line, "", true, nil
	}

	if strings.HasPrefix(line, "chat ") {
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			return "", "", "", true, fmt.Errorf("usage: chat <member> <message>")
		}
		member := strings.ToLower(strings.TrimSpace(parts[1]))
		msg := strings.TrimSpace(parts[2])
		if member == "" || msg == "" {
			return "", "", "", true, fmt.Errorf("chat requires member and message")
		}
		payload := map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": msg},
			},
		}
		b, _ := json.Marshal(payload)
		return http.MethodPost, fmt.Sprintf("/api/v1/council/%s/chat", member), string(b), true, nil
	}

	if strings.HasPrefix(line, "broadcast ") {
		msg := strings.TrimSpace(strings.TrimPrefix(line, "broadcast "))
		if msg == "" {
			return "", "", "", true, fmt.Errorf("broadcast requires a message")
		}
		payload := map[string]any{
			"content": msg,
			"source":  "cli-shell",
		}
		b, _ := json.Marshal(payload)
		return http.MethodPost, "/api/v1/swarm/broadcast", string(b), true, nil
	}

	if strings.HasPrefix(line, "send ") {
		parts := strings.SplitN(line, " ", 4)
		if len(parts) < 4 {
			return "", "", "", true, fmt.Errorf("usage: send <provider> <recipient> <message>")
		}
		payload := map[string]any{
			"provider":  parts[1],
			"recipient": parts[2],
			"message":   parts[3],
		}
		b, _ := json.Marshal(payload)
		return http.MethodPost, "/api/v1/comms/send", string(b), true, nil
	}

	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 {
		return "", "", "", false, nil
	}
	method = strings.ToUpper(strings.TrimSpace(parts[0]))
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
		target = strings.TrimSpace(parts[1])
		if len(parts) == 3 {
			body = strings.TrimSpace(parts[2])
		}
		return method, target, body, true, nil
	default:
		return "", "", "", false, nil
	}
}

func runActionShell(cfg ActionCLIConfig) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Mycelis Action Shell (base=%s)\n", cfg.APIBaseURL)
	fmt.Println("Commands: status | /path | chat <member> <message> | broadcast <message> | send <provider> <recipient> <message> | <METHOD> <PATH|URL> [JSON] | use <base_url> | help | exit")

	for {
		fmt.Print("mycelis> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			return nil
		}
		if line == "help" {
			fmt.Println("status")
			fmt.Println("/api/v1/services/status")
			fmt.Println("chat sentry health check")
			fmt.Println("broadcast summarize current mission state")
			fmt.Println("send whatsapp +15551231234 alert-text")
			fmt.Println("POST /api/v1/comms/send {\"provider\":\"slack\",\"recipient\":\"#ops\",\"message\":\"hello\"}")
			continue
		}
		if strings.HasPrefix(line, "use ") {
			newBase := strings.TrimSpace(strings.TrimPrefix(line, "use "))
			if _, err := resolveActionURL(newBase, "/healthz"); err != nil {
				fmt.Printf("Error: invalid base URL: %v\n", err)
				continue
			}
			cfg.APIBaseURL = strings.TrimRight(newBase, "/")
			fmt.Printf("Base URL set to %s\n", cfg.APIBaseURL)
			continue
		}

		method, target, body, handled, parseErr := parseActionShellLine(line)
		if parseErr != nil {
			fmt.Printf("Error: %v\n", parseErr)
			continue
		}
		if !handled {
			fmt.Println("Unrecognized command. Use `help`.")
			continue
		}

		statusCode, respBody, err := executeActionRequest(cfg, method, target, body)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		printActionResponse(statusCode, respBody)
	}
}

func runActionCLI(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: server action <METHOD> <PATH|URL> [JSON_BODY] | shell")
	}

	cfg, err := loadActionRuntimeConfig()
	if err != nil {
		return err
	}

	if strings.EqualFold(strings.TrimSpace(args[0]), "shell") {
		return runActionShell(cfg)
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: server action <METHOD> <PATH|URL> [JSON_BODY] | shell")
	}

	method := strings.ToUpper(strings.TrimSpace(args[0]))
	target := strings.TrimSpace(args[1])
	body := ""
	if len(args) > 2 {
		body = args[2]
	}

	statusCode, respBody, err := executeActionRequest(cfg, method, target, body)
	if err != nil {
		return err
	}
	printActionResponse(statusCode, respBody)

	if statusCode >= 400 {
		return fmt.Errorf("remote returned status %d", statusCode)
	}
	return nil
}
