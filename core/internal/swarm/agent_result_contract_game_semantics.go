package swarm

import (
	"regexp"
	"strings"
)

var (
	gameCanvasTextFacadePattern = regexp.MustCompile(`(?is)\b(?:game|canvas)\s*\.\s*(?:textContent|innerText|innerHTML)\s*=`)
	gameCanvasLookupPattern     = regexp.MustCompile(`(?is)(?:getElementById\s*\(\s*["']game["']\s*\)|querySelector\s*\(\s*["']#game["']\s*\))`)
	gameCanvasContextPattern    = regexp.MustCompile(`(?is)\.\s*getContext\s*\(\s*["']2d["']\s*\)`)
	gameStateModelPattern       = regexp.MustCompile(`(?is)\b(?:const|let|var)\s+(?:gameState|state|player|enemies|hazards|keys)\b|\bplayer\s*=\s*\{|\benemies\s*=\s*\[`)
	gameMovementInputPattern    = regexp.MustCompile(`(?is)(?:keydown|keyup|keyhold|ArrowLeft|ArrowRight|KeyA|KeyD|\bwasd\b)`)
	gameMovementMutationPattern = regexp.MustCompile(`(?is)(?:(?:player\s*\.\s*(?:x|y|vx|vy)|\b(?:vx|vy)\b|keys\s*\[[^]]+\])\s*(?:\+\+|--|\+=|-=|=)|keys\s*\.\s*(?:add|delete)\s*\()`)
	gameRenderedStatePattern    = regexp.MustCompile(`(?is)(?:fillRect|strokeRect|drawImage|arc|moveTo|lineTo)\s*\([^)]*(?:player\s*\.\s*(?:x|y)|\b(?:vx|vy)\b|\bx\b|\by\b)`)
	gameAttackHookPattern       = regexp.MustCompile(`(?is)\battack\w*\b`)
	gameAttackStatePattern      = regexp.MustCompile(`(?is)(?:(?:enem(?:y|ies)(?:\s*\[[^]]+\])?|\b(?:enemy|e))\s*\.\s*(?:health|hp|alive)|hazards?|score)\s*(?:\+\+|--|\+=|-=|=)|enemies\s*\.\s*(?:splice|filter)\s*\(`)
	gameHealthStatePattern      = regexp.MustCompile(`(?is)(?:player\s*\.\s*)?health\s*(?:\+\+|--|\+=|-=|=)`)
	gameKeyStatePattern         = regexp.MustCompile(`(?is)(?:hasKey|keyCollected|state\s*\.\s*(?:key|hasKey))\s*=`)
	gameScoreStatePattern       = regexp.MustCompile(`(?is)\bscore\s*(?:\+\+|--|\+=|-=|=)`)
	gameWinStatePattern         = regexp.MustCompile(`(?is)(?:gameState|state|goalState)\s*=\s*["'](?:won|win)["']`)
	gameFailStatePattern        = regexp.MustCompile(`(?is)(?:gameState|state|goalState)\s*=\s*["'](?:failed|fail|lost)["']`)
	gameRestartStatePattern     = regexp.MustCompile(`(?is)\b(?:restart|reset)\w*\b[\s\S]{0,600}(?:gameState|state|goalState|player\s*\.\s*(?:x|y|health))\s*=`)
)

func functionalGameEntrypointIssues(criteria []string, content string) []string {
	script := outputValidationJavaScript(content)
	issues := []string{}
	if gameCanvasTextFacadePattern.MatchString(script) {
		issues = append(issues, "game canvas is a text-label facade instead of a rendered validation surface")
	}
	if !gameCanvasLookupPattern.MatchString(script) || !gameCanvasContextPattern.MatchString(script) || !outputValidationDrawingEffectPattern.MatchString(script) {
		issues = append(issues, "game canvas has no inspectable 2d render implementation")
	}
	if !strings.Contains(script, "requestAnimationFrame") || len(resultContractDormantAnimationLoopIssues(content)) > 0 {
		issues = append(issues, "game entrypoint has no active canvas render or game-state loop")
	}
	if !gameStateModelPattern.MatchString(script) {
		issues = append(issues, "game entrypoint has no governed player or game-state model")
	}
	if !gameMovementInputPattern.MatchString(script) || !gameMovementMutationPattern.MatchString(script) || !gameRenderedStatePattern.MatchString(script) {
		issues = append(issues, "movement controls do not mutate game state consumed by the canvas render loop")
	}
	if hasSemanticCriterion(criteria, "attack changes enemy, hazard, or score state") &&
		(!gameAttackHookPattern.MatchString(script) || !gameAttackStatePattern.MatchString(script)) {
		issues = append(issues, "attack action does not mutate governed enemy, hazard, or score state")
	}
	if hasSemanticCriterion(criteria, "hazard contact changes health state") && !gameHealthStatePattern.MatchString(script) {
		issues = append(issues, "hazard action does not mutate governed health state")
	}
	if hasSemanticCriterion(criteria, "key pickup changes key and score state") &&
		(!gameKeyStatePattern.MatchString(script) || !gameScoreStatePattern.MatchString(script)) {
		issues = append(issues, "key action does not mutate governed key and score state")
	}
	if hasSemanticCriterion(criteria, "team play-tests the documented winning route through the locked door objective to win state") && !gameWinStatePattern.MatchString(script) {
		issues = append(issues, "win action does not transition governed game state")
	}
	if hasSemanticCriterion(criteria, "fail state can transition through restart to the initial objective") &&
		(!gameFailStatePattern.MatchString(script) || !gameRestartStatePattern.MatchString(script)) {
		issues = append(issues, "fail and restart actions do not transition governed game state")
	}
	return issues
}

func hasSemanticCriterion(criteria []string, expected string) bool {
	for _, criterion := range criteria {
		if strings.EqualFold(strings.TrimSpace(criterion), expected) {
			return true
		}
	}
	return false
}
