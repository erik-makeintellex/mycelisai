package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

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
		return parseActionChatLine(line)
	}
	if strings.HasPrefix(line, "broadcast ") {
		return parseActionBroadcastLine(line)
	}
	if strings.HasPrefix(line, "send ") {
		return parseActionSendLine(line)
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

func parseActionChatLine(line string) (method, target, body string, handled bool, err error) {
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

func parseActionBroadcastLine(line string) (method, target, body string, handled bool, err error) {
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

func parseActionSendLine(line string) (method, target, body string, handled bool, err error) {
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
			printActionShellHelp()
			continue
		}
		if strings.HasPrefix(line, "use ") {
			cfg = handleActionShellUse(cfg, line)
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

func printActionShellHelp() {
	fmt.Println("status")
	fmt.Println("/api/v1/services/status")
	fmt.Println("chat sentry health check")
	fmt.Println("broadcast summarize current mission state")
	fmt.Println("send whatsapp +15551231234 alert-text")
	fmt.Println("POST /api/v1/comms/send {\"provider\":\"slack\",\"recipient\":\"#ops\",\"message\":\"hello\"}")
}

func handleActionShellUse(cfg ActionCLIConfig, line string) ActionCLIConfig {
	newBase := strings.TrimSpace(strings.TrimPrefix(line, "use "))
	if _, err := resolveActionURL(newBase, "/healthz"); err != nil {
		fmt.Printf("Error: invalid base URL: %v\n", err)
		return cfg
	}
	cfg.APIBaseURL = strings.TrimRight(newBase, "/")
	fmt.Printf("Base URL set to %s\n", cfg.APIBaseURL)
	return cfg
}
