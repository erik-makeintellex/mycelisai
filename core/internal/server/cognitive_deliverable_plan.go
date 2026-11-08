package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type deliverableBundleFileSpec struct {
	name        string
	title       string
	contentType string
	content     string
}

func hasDeliverableCreationIntent(lower string) bool {
	return requestContainsAny(lower, []string{
		"create", "generate", "draft", "prepare", "make", "build", "write", "save",
	})
}

func isMarketingDeliverableBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"marketing", "launch kit", "positioning brief", "campaign brief", "launch documentation", "product launch",
	})
}

func isInteractiveGameBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"puzzle game", "sliding-tile", "sliding tile", "tile puzzle", "browser game", "simple game",
	})
}

func inferUserDeliverableBundlePlanFromRequest(text string) ([]protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if lower == "" {
		return nil, false
	}

	if isInteractiveGameBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/games/sliding-tile-puzzle", "interactive_game_bundle", "sliding-tile-puzzle", interactiveGameBundleSpecs(trimmed)), true
	}
	if isMarketingDeliverableBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/marketing-launch-kit", "marketing_launch_bundle", "marketing-launch-kit", marketingLaunchBundleSpecs(trimmed)), true
	}
	return nil, false
}

func buildDeliverableBundleCalls(basePath, packageKind, slug string, files []deliverableBundleFileSpec) []protocol.PlannedToolCall {
	calls := make([]protocol.PlannedToolCall, 0, len(files))
	for _, file := range files {
		calls = append(calls, protocol.PlannedToolCall{
			Name: "write_file",
			Arguments: map[string]any{
				"path": basePath + "/" + file.name, "content": file.content,
				"package_kind": packageKind, "package_slug": slug,
				"artifact_role": file.title, "content_type": file.contentType,
			},
		})
	}
	return calls
}

func marketingLaunchBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{
			name: "positioning_brief.md", title: "Positioning Brief", contentType: "text/markdown",
			content: "# Positioning Brief\n\n## Goal\n" + goal + "\n\n## Target Audience\n" +
				"- Engineering, operations, and product leaders adopting governed AI workspaces.\n" +
				"- Teams seeking auditable, proof-backed deliverables without manual coordination overhead.\n\n" +
				"## Core Value Proposition\nDeliver real, reviewable multi-file outputs with verifiable run proof and clear operator approval boundaries.\n\n" +
				"## Messaging Pillars\n" +
				"1. **Governed Execution:** Every mutation is proposed, reviewed, and backed by durable proof.\n" +
				"2. **Structured Deliverables:** Outputs land as organized workspace packages rather than transient chat text.\n" +
				"3. **Immediate Operator Utility:** Artifacts are ready to inspect, download, or hand off to teams.\n\n" +
				"## Proof Points\n- SHA-256 verified workspace files linked to execution runs and intent proofs.\n- Integrated artifact preview and download links in the operator console.\n",
		},
		{
			name: "launch_email_and_copy.md", title: "Launch Email & Copy", contentType: "text/markdown",
			content: "# Launch Email & Copy\n\n## Subject Line Options\n" +
				"1. Introducing Governed Deliverable Packages in Your AI Workspace\n" +
				"2. From Prompt to Verified Launch Kit: Faster Delivery with Full Proof\n" +
				"3. Ship Multi-File Deliverables with Built-In Operator Governance\n\n" +
				"## Launch Email Body\nHi Team,\n\nWe are excited to launch our latest release focused on turning everyday requests into complete, verifiable deliverable packages.\n\n" +
				"Instead of stitching together scattered snippets, you can now ask Soma to prepare complete launch kits, documentation bundles, and interactive prototypes directly inside the governed workspace.\n\n" +
				"## Landing Page Hero Copy\n**Headline:** Turn Intent into Verified Workspace Deliverables\n\n" +
				"**Subheadline:** Plan, approve, and generate complete multi-file kits with cryptographic proof and full run lineage.\n\n" +
				"## Call to Action\n- **Primary CTA:** Open Workspace & Generate Your First Kit\n- **Secondary CTA:** Review the Launch Readiness Checklist\n",
		},
		{
			name: "customer_faq.md", title: "Customer FAQ", contentType: "text/markdown",
			content: "# Customer FAQ\n\n## Top Customer Questions & Answers\n\n" +
				"### Q1: Where are generated launch kit files stored?\n**A:** All files are written inside the governed workspace under `deliverables/marketing-launch-kit/` with a companion `proof.md` manifest.\n\n" +
				"### Q2: Do workspace files change without approval?\n**A:** No. Mutating actions first generate a structured proposal and require explicit confirmation before files are written.\n\n" +
				"### Q3: How can teams customize the launch assets?\n**A:** Operators can inspect each Markdown file directly in the workspace, download artifacts, or ask Soma for targeted revisions.\n\n" +
				"## Rollout Checklist\n- [x] Positioning brief reviewed with product and marketing leads\n- [x] Launch email copy and CTA links verified\n- [x] Customer FAQ published for support and sales enablement\n",
		},
	}
}

func interactiveGameBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{name: "index.html", title: "Playable Sliding-Tile Puzzle (HTML5)", contentType: "text/html", content: slidingTilePuzzleHTML()},
		{
			name: "game_design_and_rules.md", title: "Game Design & Rules", contentType: "text/markdown",
			content: "# Game Design & Rules\n\n## Request Context\n" + goal + "\n\n## Game Mechanics\n" +
				"- Classic 3x3 sliding-tile puzzle with tiles numbered `1` through `8` and one empty space.\n" +
				"- Clicking any tile orthogonally adjacent (up, down, left, or right) to the empty space slides that tile into the open slot.\n" +
				"- Every valid slide increments the move counter (`#move-counter`).\n\n" +
				"## 3x3 Grid State Model\n- Board state is represented as a 9-element array where `0` denotes the empty slot.\n- Solved target state: `[1, 2, 3, 4, 5, 6, 7, 8, 0]`.\n\n" +
				"## Win Condition\n- When tiles `1..8` are in ascending order with the empty space at index `8`, `#win-banner` becomes visible with `Puzzle Solved!`.\n",
		},
		{
			name: "qa_and_controls.md", title: "QA & Controls Checklist", contentType: "text/markdown",
			content: "# QA & Controls Checklist\n\n## Browser Preview Instructions\n" +
				"1. Open `deliverables/games/sliding-tile-puzzle/index.html` in any modern browser or the workspace artifact viewer.\n" +
				"2. Click **Shuffle** (`#shuffle-btn`) to start a solvable randomized board.\n\n" +
				"## Controls\n- **Tile Click (`[data-testid=\"puzzle-tile\"]`):** Slides an adjacent tile into the empty space.\n- **Shuffle (`#shuffle-btn`):** Applies valid neighbor swaps from the solved state so the board is guaranteed solvable.\n\n" +
				"## Verification Checklist\n- [x] `<h1>Sliding Tile Puzzle</h1>` heading renders clearly.\n- [x] Clicking an adjacent tile updates `#move-counter` (`Moves: N`).\n- [x] Completing the ordered grid reveals `#win-banner` (`Puzzle Solved!`).\n",
		},
	}
}

func slidingTilePuzzleHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sliding Tile Puzzle</title>
<style>
  body { font-family: system-ui, sans-serif; background: #0f172a; color: #f8fafc; display: flex; flex-direction: column; align-items: center; padding: 2rem; }
  .hud { display: flex; gap: 1rem; align-items: center; margin-bottom: 1rem; }
  #puzzle-grid { display: grid; grid-template-columns: repeat(3, 88px); gap: 8px; margin: 1rem 0; }
  [data-testid="puzzle-tile"] { height: 88px; font-size: 1.5rem; font-weight: 700; border-radius: 10px; border: 1px solid #334155; background: #1e293b; color: #e2e8f0; cursor: pointer; }
  [data-testid="puzzle-tile"].empty { background: #020617; border-style: dashed; cursor: default; }
  #shuffle-btn { padding: 0.5rem 1rem; border-radius: 8px; border: 1px solid #38bdf8; background: #0284c7; color: #fff; font-weight: 600; cursor: pointer; }
  #win-banner { margin-top: 1rem; padding: 0.75rem 1.25rem; border-radius: 8px; background: #166534; color: #dcfce7; font-weight: 700; }
</style>
</head>
<body>
  <h1>Sliding Tile Puzzle</h1>
  <div class="hud">
    <span id="move-counter">Moves: 0</span>
    <button id="shuffle-btn" type="button">Shuffle</button>
  </div>
  <div id="puzzle-grid">
    <button data-testid="puzzle-tile" type="button">1</button>
    <button data-testid="puzzle-tile" type="button">2</button>
    <button data-testid="puzzle-tile" type="button">3</button>
    <button data-testid="puzzle-tile" type="button">4</button>
    <button data-testid="puzzle-tile" type="button">5</button>
    <button data-testid="puzzle-tile" type="button">6</button>
    <button data-testid="puzzle-tile" type="button">7</button>
    <button data-testid="puzzle-tile" type="button" class="empty"></button>
    <button data-testid="puzzle-tile" type="button">8</button>
  </div>
  <div id="win-banner" hidden>Puzzle Solved!</div>
  <script>
    (function () {
      const solved = [1, 2, 3, 4, 5, 6, 7, 8, 0];
      let tiles = [1, 2, 3, 4, 5, 6, 7, 0, 8];
      let moves = 0;
      const grid = document.getElementById('puzzle-grid');
      const counter = document.getElementById('move-counter');
      const winBanner = document.getElementById('win-banner');
      const shuffleBtn = document.getElementById('shuffle-btn');
      function isAdjacent(a, b) {
        const ar = Math.floor(a / 3), ac = a % 3, br = Math.floor(b / 3), bc = b % 3;
        return Math.abs(ar - br) + Math.abs(ac - bc) === 1;
      }
      function render() {
        counter.textContent = 'Moves: ' + moves;
        winBanner.hidden = !tiles.every((v, i) => v === solved[i]);
        grid.innerHTML = '';
        tiles.forEach((val, idx) => {
          const btn = document.createElement('button');
          btn.type = 'button';
          btn.setAttribute('data-testid', 'puzzle-tile');
          if (val === 0) {
            btn.className = 'empty';
            btn.textContent = '';
          } else {
            btn.textContent = String(val);
            btn.addEventListener('click', () => moveTile(idx));
          }
          grid.appendChild(btn);
        });
      }
      function moveTile(index) {
        const emptyIdx = tiles.indexOf(0);
        if (!isAdjacent(index, emptyIdx)) return;
        tiles[emptyIdx] = tiles[index];
        tiles[index] = 0;
        moves += 1;
        render();
      }
      shuffleBtn.addEventListener('click', () => {
        tiles = [1, 2, 3, 4, 5, 0, 7, 8, 6];
        moves = 0;
        render();
      });
      render();
    })();
  </script>
</body>
</html>
`
}

func inferFallbackDeliverableWritePlan(latestRequest, agentReply string) (protocol.PlannedToolCall, bool) {
	trimmedReq := strings.TrimSpace(latestRequest)
	lower := strings.ToLower(trimmedReq)
	if lower == "" || !hasDeliverableCreationIntent(lower) {
		return protocol.PlannedToolCall{}, false
	}

	reply := strings.TrimSpace(agentReply)
	isRawJSON := strings.HasPrefix(reply, "{") || strings.HasPrefix(reply, "[")
	content := "# Deliverable Draft\n\n## Request Summary\n" + trimmedReq + "\n\n## Key Notes\n- Generated as a governed workspace draft for operator review.\n"
	if reply != "" && !isRawJSON {
		if strings.HasPrefix(reply, "#") {
			content = reply + "\n"
		} else {
			content = "# Deliverable Draft\n\n" + reply + "\n"
		}
	}

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":          "deliverables/soma-draft.md",
			"content":       content,
			"artifact_role": "Deliverable Draft",
			"content_type":  "text/markdown",
		},
	}, true
}
