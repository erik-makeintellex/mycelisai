package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/hostcmd"
	"github.com/mycelis/core/pkg/protocol"
)

func (r *InternalToolRegistry) handleBroadcast(_ context.Context, args map[string]any) (string, error) {
	message := stringValue(args["message"])
	if message == "" {
		return "", fmt.Errorf("broadcast requires 'message'")
	}
	if urgency := stringValue(args["urgency"]); urgency != "" {
		log.Printf("Broadcast urgency: %s", urgency)
	}
	if r.nc == nil {
		return "", fmt.Errorf("NATS not available")
	}
	if r.somaRef == nil {
		return "", fmt.Errorf("Soma not available — cannot enumerate teams")
	}
	teams := r.somaRef.ListTeams()
	if len(teams) == 0 {
		return "No active teams to broadcast to.", nil
	}

	type reply struct {
		teamID  string
		content string
		err     error
	}
	ch := make(chan reply, len(teams))
	for _, t := range teams {
		go func(teamID string) {
			msg, err := r.nc.Request(fmt.Sprintf(protocol.TopicTeamInternalTrigger, teamID), []byte(message), 60*time.Second)
			if err != nil {
				ch <- reply{teamID: teamID, err: err}
				return
			}
			ch <- reply{teamID: teamID, content: string(msg.Data)}
		}(t.ID)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Broadcast sent to %d team(s):\n\n", len(teams)))
	for range teams {
		reply := <-ch
		if reply.err != nil {
			sb.WriteString(fmt.Sprintf("- **%s**: _%v_\n", reply.teamID, reply.err))
		} else {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", reply.teamID, reply.content))
		}
	}
	return sb.String(), nil
}

func (r *InternalToolRegistry) handleSendExternalMessage(ctx context.Context, args map[string]any) (string, error) {
	if r.comms == nil {
		return "", fmt.Errorf("communications gateway unavailable")
	}
	provider := stringValue(args["provider"])
	message := stringValue(args["message"])
	if provider == "" || message == "" {
		return "", fmt.Errorf("send_external_message requires provider and message")
	}
	metadata, _ := args["metadata"].(map[string]any)
	res, err := r.comms.Send(ctx, comms.SendRequest{Provider: provider, Recipient: stringValue(args["recipient"]), Message: message, Metadata: metadata})
	if err != nil {
		return "", fmt.Errorf("external send failed: %w", err)
	}
	return mustJSON(map[string]any{"message": fmt.Sprintf("external message sent via %s", res.Provider), "provider": res.Provider, "status": res.Status, "result": res}), nil
}

func (r *InternalToolRegistry) handleReadFile(_ context.Context, args map[string]any) (string, error) {
	path := stringValue(args["path"])
	if path == "" {
		return "", fmt.Errorf("read_file requires 'path'")
	}
	safePath, err := validateToolPath(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(safePath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", safePath, err)
	}
	content := string(data)
	if len(content) > 32000 {
		content = content[:32000] + "\n... [truncated at 32KB]"
	}
	return content, nil
}

func (r *InternalToolRegistry) handleWriteFile(_ context.Context, args map[string]any) (string, error) {
	path := stringValue(args["path"])
	content := stringValue(args["content"])
	if path == "" || content == "" {
		return "", fmt.Errorf("write_file requires 'path' and 'content'")
	}
	if len(content) > maxWriteSize {
		return "", fmt.Errorf("content size %d exceeds maximum write size of %d bytes", len(content), maxWriteSize)
	}
	safePath, err := validateToolPath(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(safePath), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", filepath.Dir(safePath), err)
	}
	if err := os.WriteFile(safePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", safePath, err)
	}
	supportCount, err := writeProjectPackageSupportFiles(path, args)
	if err != nil {
		return "", err
	}
	if supportCount > 0 {
		return fmt.Sprintf("File written: %s (%d bytes). Project package support files written: %d.", safePath, len(content), supportCount), nil
	}
	return fmt.Sprintf("File written: %s (%d bytes).", safePath, len(content)), nil
}

func writeProjectPackageSupportFiles(mainPath string, args map[string]any) (int, error) {
	if !strings.EqualFold(strings.TrimSpace(stringValue(args["package_kind"])), "project_package") {
		return 0, nil
	}
	folder := strings.TrimSpace(stringValue(args["package_folder"]))
	if folder == "" {
		folder = filepath.Dir(normalizeWorkspaceRelativePath(mainPath))
	}
	if folder == "." || folder == "" {
		return 0, nil
	}
	safeFolder, err := validateToolPath(folder)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(safeFolder, 0o755); err != nil {
		return 0, fmt.Errorf("failed to create package folder %s: %w", safeFolder, err)
	}

	count := 0
	for _, file := range stringSlice(args["package_files"]) {
		rel := strings.Trim(strings.TrimSpace(file), `/\`)
		if rel == "" {
			continue
		}
		base := strings.ToLower(filepath.Base(rel))
		if base != "readme.md" && base != "proof.md" && base != "validation-notes.md" {
			continue
		}
		cleanRel := filepath.Clean(filepath.FromSlash(rel))
		if filepath.IsAbs(cleanRel) || cleanRel == "." || strings.HasPrefix(cleanRel, "..") {
			return count, fmt.Errorf("package support file %q escapes package folder", file)
		}
		target := filepath.Join(safeFolder, cleanRel)
		relToFolder, err := filepath.Rel(safeFolder, target)
		if err != nil || strings.HasPrefix(relToFolder, "..") {
			return count, fmt.Errorf("package support file %q escapes package folder", file)
		}
		content := projectPackageSupportFileContent(base, args, mainPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return count, fmt.Errorf("failed to create package support folder %s: %w", filepath.Dir(target), err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return count, fmt.Errorf("failed to write package support file %s: %w", target, err)
		}
		count++
	}
	return count, nil
}

func projectPackageSupportFileContent(file string, args map[string]any, mainPath string) string {
	title := strings.TrimSpace(stringValue(args["package_title"]))
	if title == "" {
		title = "Generated project package"
	}
	entrypoint := strings.TrimSpace(stringValue(args["package_entrypoint"]))
	if entrypoint == "" {
		entrypoint = mainPath
	}
	validation := strings.TrimSpace(stringValue(args["validation"]))
	if validation == "" {
		validation = strings.TrimSpace(stringValue(args["validation_summary"]))
	}
	if validation == "" {
		validation = "Open the entrypoint in a browser and review the retained output."
	}
	if file == "proof.md" || file == "validation-notes.md" {
		return fmt.Sprintf("# %s Proof\n\n- Entrypoint: `%s`\n- Validation: %s\n- Generated graphics/assets: code-only browser output.\n- Recovery: rerun or ask Soma for a revision if browser validation fails.\n", title, entrypoint, validation)
	}
	return fmt.Sprintf("# %s\n\n## Open\n\nOpen `%s` in a browser.\n\n## Controls\n\nUse WASD or arrow keys to move. Press `R` or `Restart` to restart.\n\n## Validation\n\n%s\n", title, entrypoint, validation)
}

func (r *InternalToolRegistry) handleLocalCommand(ctx context.Context, args map[string]any) (string, error) {
	command := stringValue(args["command"])
	if command == "" {
		return "", fmt.Errorf("local_command requires 'command'")
	}
	cmdArgs := stringSlice(args["args"])
	timeoutMS := 5000
	if v, ok := args["timeout_ms"].(float64); ok {
		timeoutMS = int(v)
	}
	if len(cmdArgs) == 0 && strings.ContainsAny(command, " \t\r\n'\"`|&;<>") {
		return "", fmt.Errorf("local_command requires a bare allowlisted command name in 'command' and separate 'args'; shell snippets are not allowed")
	}
	result, err := hostcmd.Execute(ctx, command, cmdArgs, time.Duration(timeoutMS)*time.Millisecond)
	if err != nil {
		return "", err
	}
	payload, _ := json.Marshal(result)
	return string(payload), nil
}
