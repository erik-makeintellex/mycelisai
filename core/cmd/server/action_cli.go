package main

import (
	"fmt"
	"strings"
)

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
