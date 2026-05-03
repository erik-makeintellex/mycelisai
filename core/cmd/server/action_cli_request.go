package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
