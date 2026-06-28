package server

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

func generatedTeamIDForRequest(name, lower string) string {
	base := slugID(name)
	if base == "" || base == "soma-requested-team" {
		base = generatedTeamIDBase(lower)
	}
	prefix := generatedTeamIDPrefix(lower)
	if prefix != "" {
		base = strings.TrimPrefix(base, "temporary-")
		base = strings.TrimPrefix(base, "temp-")
		base = strings.TrimPrefix(base, "standing-")
		if !strings.HasPrefix(base, prefix+"-") {
			base = prefix + "-" + strings.Trim(base, "-")
		}
	}
	base = strings.Trim(base, "-")
	if base == "" {
		base = "soma-team"
	}
	return base + "-" + shortStableSuffix(name+"\n"+lower)
}

func generatedTeamNameForRequest(lower string) string {
	base := generatedTeamNameBase(lower)
	switch generatedTeamIDPrefix(lower) {
	case "temp":
		if strings.HasPrefix(base, "Temporary ") {
			return base
		}
		return "Temporary " + base
	case "standing":
		if strings.HasPrefix(base, "Standing ") {
			return base
		}
		return "Standing " + base
	default:
		return base
	}
}

func generatedTeamNameBase(lower string) string {
	matches := 0
	if strings.Contains(lower, "game") {
		matches++
	}
	if requestAsksForMedia(lower) {
		matches++
	}
	if requestAsksForTextOutput(lower) {
		matches++
	}
	if matches > 1 {
		return "Mixed Output Team"
	}
	switch {
	case strings.Contains(lower, "game"):
		return "Game Delivery Team"
	case requestAsksForMedia(lower):
		return "Media Generation Team"
	case requestContainsAny(lower, []string{"watch", "watcher", "monitor", "steward", "react to", "reaction"}):
		return "Content Steward Team"
	case strings.Contains(lower, "research"):
		return "AI Research Team"
	case requestAsksForTextOutput(lower):
		return "Writing Delivery Team"
	default:
		return "Soma Delivery Team"
	}
}

func generatedTeamIDBase(lower string) string {
	matches := 0
	if strings.Contains(lower, "game") {
		matches++
	}
	if requestAsksForMedia(lower) {
		matches++
	}
	if requestAsksForTextOutput(lower) {
		matches++
	}
	if matches > 1 {
		return "mixed-output-team"
	}
	switch {
	case strings.Contains(lower, "game"):
		return "game-team"
	case requestAsksForMedia(lower):
		return "media-team"
	case strings.Contains(lower, "research"):
		return "research-team"
	default:
		return "soma-team"
	}
}

func generatedTeamIDPrefix(lower string) string {
	switch {
	case requestContainsAny(lower, []string{"temporary", "temp team", "temp-team"}):
		return "temp"
	case requestContainsAny(lower, []string{"standing", "consistent", "watch", "watcher", "monitor"}):
		return "standing"
	default:
		return ""
	}
}

func shortStableSuffix(text string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(strings.ToLower(text))))
	return fmt.Sprintf("%x", sum)[:5]
}
