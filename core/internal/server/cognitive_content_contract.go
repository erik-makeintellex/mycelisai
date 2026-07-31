package server

import "strings"

func contentContractForTeamRequest(request string) map[string]any {
	lower := strings.ToLower(strings.TrimSpace(request))
	types := []string{}
	criteria := []string{}
	outputs := []string{}
	proof := []string{"retained output path", "operator-readable summary", "recovery note if incomplete"}
	preparation := []string{}
	if strings.Contains(lower, "game") {
		types = append(types, "game")
		outputs = append(outputs, "openable browser game package")
		criteria = append(criteria, gameAcceptanceCriteria()...)
		proof = append(proof, gameProofRequirements()...)
	}
	if requestAsksForTableOrData(lower) {
		types = append(types, "table_data")
		outputs = append(outputs, "structured table or data report")
		criteria = append(criteria,
			"columns and rows match the operator's comparison or data-management goal",
			"source boundary, assumptions, and missing data are separated from table rows",
			"table can be read in Soma chat and retained as a reviewable file when requested",
		)
		proof = append(proof, "table schema or column rationale", "source/assumption notes")
	}
	if requestAsksForApplicationPackage(lower) && !strings.Contains(lower, "game") {
		types = append(types, "application_package")
		outputs = append(outputs, "openable application package")
		criteria = append(criteria, applicationPackageAcceptanceCriteria()...)
		proof = append(proof, applicationPackageProofRequirements()...)
	}
	if requestAsksForCodeArtifact(lower) && !strings.Contains(lower, "game") {
		types = append(types, "code_app")
		outputs = append(outputs, "reviewable code or script package")
		criteria = append(criteria,
			"code/script package matches the requested deployment target",
			"launch or usage path is provided for the operator or another agent",
			"validation notes state what was run or what remains unverified",
		)
		proof = append(proof, "launch or build instructions", "code/package validation notes")
	}
	if requestAsksForMedia(lower) {
		types = append(types, "media")
		outputs = append(outputs, "saved media artifact")
		criteria = append(criteria,
			"prompt reflects requested subject, style, and constraints",
			"artifact is saved to a governed workspace path",
			"media boundary/provider status is clear",
			"visual review notes identify whether the result matches the ask",
		)
		proof = append(proof, "media artifact path", "generation/provider proof")
	}
	if requestAsksForTextOutput(lower) {
		types = append(types, "text")
		outputs = append(outputs, "readable retained text file")
		criteria = append(criteria,
			"text directly answers the requested job",
			"structure matches the requested format",
			"claims, sources, or assumptions are separated when relevant",
			"file can be reopened from Resources or the Outcome",
		)
		proof = append(proof, "file readback proof")
	}
	if len(types) == 0 {
		types = append(types, "work_product")
		outputs = append(outputs, "retained work result")
		criteria = append(criteria, "result matches the operator request", "next action is clear")
	}
	if requestNeedsTeamPreparation(lower, types) {
		preparation = append(preparation,
			"research available external/current context or local sources before finalizing team roles",
			"consult council or equivalent review before choosing implementation stack, roles, and proof gates",
			"state the chosen output format and why it fits the operator's deployment target",
			"identify specialist additions only with owned tasks, proof expected, and removal point",
		)
		proof = append(proof, "research/council preparation summary when complex delivery is requested")
	}
	return map[string]any{
		"content_types":       types,
		"expected_outputs":    uniqueOrderedTools(outputs),
		"acceptance_criteria": uniqueOrderedTools(criteria),
		"proof_required":      uniqueOrderedTools(proof),
		"team_preparation":    uniqueOrderedTools(preparation),
	}
}

func requestNeedsTeamPreparation(lower string, types []string) bool {
	if requestAsksTeamToDeliver(lower) {
		return true
	}
	if len(types) == 1 && types[0] == "work_product" {
		return requestContainsAny(lower, []string{
			"complex", "advanced", "production", "deploy", "deployable", "executable", "application", "app", "package",
			"internet", "look up", "latest", "research on", "research the", "research available", "team leads", "specialists", "architecture", "multi-team",
		})
	}
	return len(types) > 1 || requestContainsAny(lower, []string{
		"complex", "advanced", "production", "deploy", "deployable", "executable", "application", "app", "package",
		"detailed", "substantial", "hard", "research", "internet", "look up", "latest", "team leads", "specialists", "architecture", "multi-team",
	})
}

func gameAcceptanceCriteria() []string {
	return []string{
		"visible in-app instructions name the primary control and objective",
		"playable controls respond in browser",
		"visible game loop renders without a blank canvas",
		"collision or boundary rules affect play",
		"objective, win/fail state, and restart are testable",
		"known winning route or walkthrough is documented",
		"team play-tests the route from start to win before delivery",
		"validation defects are reported back through Soma as a repair request before direct output edits",
		"direct launch or view path is provided for the user or another agent",
		"matching music or action audio exists when the game request asks for media or sound",
		"controls, objective, and recovery/restart instructions are documented",
	}
}

func gameProofRequirements() []string {
	return []string{
		"headed gameplay proof or equivalent interaction proof",
		"play-through notes or screenshots showing start, objective pickup, win, and restart",
		"Soma repair-turn transcript when validation finds defects",
		"chat, Outcome, or Resources launch reference for opening the generated content",
		"audio unlock/playback check when sound or music is part of the request",
	}
}

func gameValidationSummary() string {
	return "Retained as a self-contained browser adventure with movement, collision, hazards, enemies, key, door, win/fail states, restart, matching generated music/action audio, documented winning route, play-tested route proof from start to win, Soma-mediated repair notes for any discovered defect, and a direct launch path for the user or another agent."
}

func applicationPackageAcceptanceCriteria() []string {
	return []string{
		"direct open or launch path is provided in chat and retained output metadata",
		"the primary control or workflow is visible, understandable, and changes application state when used",
		"primary user workflows are named and validated, including create/read/update or equivalent interactions when requested",
		"data/table views render with readable columns, empty states, and source or assumption boundaries when relevant",
		"package includes operator-readable usage notes and any setup limits",
		"browser or runtime validation notes state console/runtime errors, unsupported capabilities, and remaining uncertainty",
		"follow-up changes can be requested through Soma against the same Outcome or team output",
	}
}

func applicationPackageProofRequirements() []string {
	return []string{
		"application entrypoint or launch reference",
		"headed browser or runtime interaction proof for the primary workflow",
		"file/folder readback proof for retained package contents",
		"validation report with pass/fail findings and repair notes",
		"chat, Outcome, or Resources reference for reopening the generated application",
	}
}

func requestAsksForTextOutput(lower string) bool {
	if strings.TrimSpace(lower) == "" {
		return false
	}
	return requestContainsAny(lower, []string{
		"text", "markdown", "md file", "document", "doc", "report", "brief", "plan", "readme", "copy", "article", "notes",
	})
}

func requestAsksForTableOrData(lower string) bool {
	if requestContainsAny(lower, []string{"data management", "data table", "data review", "structured data"}) {
		return true
	}
	for _, word := range []string{"table", "spreadsheet", "csv", "matrix", "columns", "rows", "dataset"} {
		if hasExactWord(lower, word) {
			return true
		}
	}
	return false
}

func requestAsksForApplicationPackage(lower string) bool {
	if requestContainsAny(lower, []string{"browser app", "web app", "mobile app"}) {
		return true
	}
	for _, word := range []string{"app", "application", "executable", "package", "deployable", "tool"} {
		if hasExactWord(lower, word) {
			return true
		}
	}
	return false
}

func requestAsksForCodeArtifact(lower string) bool {
	if requestContainsAny(lower, []string{"source code"}) {
		return true
	}
	for _, word := range []string{"code", "script", "repository"} {
		if hasExactWord(lower, word) {
			return true
		}
	}
	return false
}

func requiredCapabilitiesForContentContract(contract map[string]any) []string {
	capabilities := []string{"team_orchestration"}
	for _, kind := range confirmedActionStringSlice(contract["content_types"]) {
		switch kind {
		case "game", "text", "table_data", "application_package", "code_app":
			capabilities = append(capabilities, "write_file", "store_artifact")
		case "media":
			capabilities = append(capabilities, "generate_image", "save_cached_image", "store_artifact")
		}
	}
	if contentContractNeedsResearch(contract) {
		capabilities = append(capabilities, "research_for_blueprint", "consult_council")
	}
	return uniqueOrderedTools(capabilities)
}
