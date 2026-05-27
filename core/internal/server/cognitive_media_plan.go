package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func inferTeamMediaDeliverablePlanFromRequest(text string, teamCall protocol.PlannedToolCall) (protocol.PlannedToolCall, protocol.PlannedToolCall, bool) {
	if !requestAsksForMedia(text) {
		return protocol.PlannedToolCall{}, protocol.PlannedToolCall{}, false
	}
	teamID := firstNonEmptyString(teamCall.Arguments["team_id"], teamCall.Arguments["id"], teamCall.Arguments["team_name"])
	teamName := firstNonEmptyString(teamCall.Arguments["name"], teamID, "Soma Media Team")
	return mediaDeliverableCalls(text, teamID, teamName)
}

func inferStandaloneMediaDeliverablePlanFromRequest(text string) (protocol.PlannedToolCall, protocol.PlannedToolCall, bool) {
	if !requestAsksForMedia(text) {
		return protocol.PlannedToolCall{}, protocol.PlannedToolCall{}, false
	}
	return mediaDeliverableCalls(text, "", "Soma Media Output")
}

func requestAsksForMedia(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	mediaTargets := []string{"image", "images", "picture", "illustration", "comic", "comic book", "media", "artwork", "visual"}
	return requestContainsAny(lower, mediaTargets)
}

func mediaDeliverableCalls(text, teamID, titleSeed string) (protocol.PlannedToolCall, protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return protocol.PlannedToolCall{}, protocol.PlannedToolCall{}, false
	}
	slug := slugID(firstNonEmptyString(teamID, titleSeed, "soma-media-output"))
	if slug == "" {
		slug = "soma-media-output"
	}
	prompt := mediaPromptForRequest(trimmed)
	imageArgs := map[string]any{
		"prompt":      prompt,
		"size":        mediaSizeForRequest(trimmed),
		"goal":        trimmed,
		"description": "Generate the requested media artifact with the configured local/private media provider.",
	}
	saveArgs := map[string]any{
		"folder":      "saved-media",
		"filename":    slug + ".png",
		"goal":        "Persist generated media output to the governed workspace saved-media folder.",
		"description": "Save the latest cached generated image so the operator can open the local output folder.",
	}
	if strings.TrimSpace(teamID) != "" {
		imageArgs["team_id"] = teamID
		saveArgs["team_id"] = teamID
		if groupFolder := groupWorkspaceFolderForTeamID(teamID); groupFolder != "" {
			saveArgs["folder"] = groupFolder + "/media"
			saveArgs["goal"] = "Persist generated media output to the team's dedicated group workspace folder."
		}
	}
	return protocol.PlannedToolCall{Name: "generate_image", Arguments: imageArgs},
		protocol.PlannedToolCall{Name: "save_cached_image", Arguments: saveArgs}, true
}

func mediaPromptForRequest(request string) string {
	lower := strings.ToLower(request)
	if strings.Contains(lower, "comic") {
		return "Create a polished vertical comic book page from this operator request. Use clear panel composition, readable visual storytelling, expressive characters, and leave speech balloon space for lettering. Local/private generation only. Request: " + request
	}
	return "Create a polished visual media artifact from this operator request. Local/private generation only. Request: " + request
}

func mediaSizeForRequest(request string) string {
	lower := strings.ToLower(request)
	if strings.Contains(lower, "comic") || strings.Contains(lower, "page") || strings.Contains(lower, "poster") {
		return "768x1024"
	}
	return "1024x1024"
}
