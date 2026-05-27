package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func inferFirstTeamGameDeliverablePlanFromRequest(text string, teamCall protocol.PlannedToolCall) (protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if trimmed == "" || !strings.Contains(lower, "game") {
		return protocol.PlannedToolCall{}, false
	}
	if !requestContainsAny(lower, []string{"work on", "build", "develop", "create", "make", "prototype", "playable", "detailed", "deliver", "start"}) {
		return protocol.PlannedToolCall{}, false
	}

	teamID := firstNonEmptyString(teamCall.Arguments["team_id"], teamCall.Arguments["id"], teamCall.Arguments["team_name"])
	teamName := firstNonEmptyString(teamCall.Arguments["name"], teamID, "Soma Game Team")
	slug := slugID(firstNonEmptyString(teamID, teamName, "soma-game-team"))
	if slug == "" {
		slug = "soma-game-team"
	}
	folder := "workspace/generated/" + slug + "-first-game"
	if teamID != "" {
		if groupFolder := groupWorkspaceFolderForTeamID(teamID); groupFolder != "" {
			folder = groupFolder + "/generated/first-game"
		}
	}
	entrypoint := folder + "/index.html"
	title := firstNonEmptyString(teamName, "Soma Game Team") + " First Playable"

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":               entrypoint,
			"content":            firstTeamGameHTML(title, trimmed),
			"package_kind":       "project_package",
			"package_title":      title,
			"package_folder":     folder,
			"package_entrypoint": entrypoint,
			"package_files":      firstTeamGamePackageFiles(trimmed),
			"validation":         "Retained as a self-contained browser game output for operator review.",
		},
	}, true
}

func firstTeamGamePackageFiles(request string) []string {
	files := []string{"index.html"}
	if projectPackageTextRequestsReadme(request) {
		files = append(files, "README.md")
	}
	return files
}

func firstTeamGameHTML(title, request string) string {
	escape := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	replacer := strings.NewReplacer(
		"__TITLE__", escape.Replace(firstNonEmptyString(title, "Soma Game First Playable")),
		"__REQUEST__", escape.Replace(request),
	)
	return replacer.Replace(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>__TITLE__</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, system-ui, sans-serif; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: #0c1220; color: #eef4ff; }
    main { width: min(960px, calc(100vw - 32px)); display: grid; gap: 16px; }
    header { display: flex; justify-content: space-between; gap: 12px; align-items: end; }
    h1 { margin: 0; font-size: 28px; }
    p { margin: 4px 0 0; color: #aab8d4; }
    .hud { display: flex; flex-wrap: wrap; gap: 8px; }
    .pill { border: 1px solid #31405f; border-radius: 6px; padding: 8px 10px; background: #121b2e; font-weight: 700; }
    canvas { width: 100%; aspect-ratio: 16 / 9; border: 1px solid #31405f; border-radius: 8px; background: #101827; }
    .controls { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; color: #aab8d4; }
    button { border: 1px solid #65d6b4; border-radius: 6px; background: #14352f; color: #c9fff1; padding: 10px 14px; font-weight: 800; cursor: pointer; }
  </style>
</head>
<body>
  <main>
    <header>
      <div>
        <h1>__TITLE__</h1>
        <p>First retained team deliverable from: __REQUEST__</p>
      </div>
      <div class="hud">
        <div class="pill">Score <span id="score">0</span></div>
        <div class="pill">Health <span id="health">3</span></div>
        <div class="pill">Stage <span id="stage">1</span></div>
      </div>
    </header>
    <canvas id="game" width="960" height="540" aria-label="Playable browser game"></canvas>
    <div class="controls">
      <button id="restart">Restart</button>
      <span>Move with arrows or WASD. Collect bright cores, avoid red drones, reach the beacon.</span>
    </div>
  </main>
  <script>
    const canvas = document.getElementById("game");
    const ctx = canvas.getContext("2d");
    const scoreEl = document.getElementById("score");
    const healthEl = document.getElementById("health");
    const stageEl = document.getElementById("stage");
    const keys = new Set();
    const player = { x: 60, y: 260, r: 15, vx: 0, vy: 0 };
    let score = 0;
    let health = 3;
    let stage = 1;
    let tick = 0;
    let cores = [];
    let drones = [];
    let won = false;

    function reset() {
      player.x = 60; player.y = 260; player.vx = 0; player.vy = 0;
      score = 0; health = 3; stage = 1; tick = 0; won = false;
      cores = Array.from({ length: 10 }, (_, i) => ({ x: 170 + i * 72, y: 100 + (i % 4) * 90, taken: false }));
      drones = Array.from({ length: 5 }, (_, i) => ({ x: 280 + i * 120, y: 80 + i * 70, phase: i * 40 }));
      syncHud();
    }

    function syncHud() {
      scoreEl.textContent = score;
      healthEl.textContent = health;
      stageEl.textContent = stage;
    }

    function update() {
      tick += 1;
      const ax = (keys.has("ArrowRight") || keys.has("d") ? 0.55 : 0) - (keys.has("ArrowLeft") || keys.has("a") ? 0.55 : 0);
      const ay = (keys.has("ArrowDown") || keys.has("s") ? 0.55 : 0) - (keys.has("ArrowUp") || keys.has("w") ? 0.55 : 0);
      player.vx = (player.vx + ax) * 0.88;
      player.vy = (player.vy + ay) * 0.88;
      player.x = Math.max(player.r, Math.min(canvas.width - player.r, player.x + player.vx));
      player.y = Math.max(player.r, Math.min(canvas.height - player.r, player.y + player.vy));

      for (const core of cores) {
        if (!core.taken && Math.hypot(player.x - core.x, player.y - core.y) < player.r + 13) {
          core.taken = true;
          score += 10;
          if (score % 40 === 0) stage += 1;
        }
      }
      for (const drone of drones) {
        const dx = Math.sin((tick + drone.phase) / 42) * 105;
        const dy = Math.cos((tick + drone.phase) / 35) * 48;
        drone.cx = drone.x + dx;
        drone.cy = drone.y + dy;
        if (Math.hypot(player.x - drone.cx, player.y - drone.cy) < player.r + 16 && tick % 34 === 0) {
          health -= 1;
        }
      }
      if (score >= 100 && player.x > canvas.width - 86) won = true;
      if (health <= 0) reset();
      syncHud();
    }

    function draw() {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      const gradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height);
      gradient.addColorStop(0, "#12213b");
      gradient.addColorStop(1, "#0b1020");
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.strokeStyle = "#263858";
      for (let x = 0; x < canvas.width; x += 48) { ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, canvas.height); ctx.stroke(); }
      for (let y = 0; y < canvas.height; y += 48) { ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(canvas.width, y); ctx.stroke(); }
      ctx.fillStyle = "#65d6b4";
      ctx.fillRect(canvas.width - 46, 205, 18, 130);
      for (const core of cores) {
        if (core.taken) continue;
        ctx.beginPath(); ctx.fillStyle = "#f7d774"; ctx.arc(core.x, core.y, 12, 0, Math.PI * 2); ctx.fill();
      }
      for (const drone of drones) {
        ctx.beginPath(); ctx.fillStyle = "#ff6b7a"; ctx.arc(drone.cx, drone.cy, 16, 0, Math.PI * 2); ctx.fill();
      }
      ctx.beginPath(); ctx.fillStyle = "#79a7ff"; ctx.arc(player.x, player.y, player.r, 0, Math.PI * 2); ctx.fill();
      if (won) {
        ctx.fillStyle = "rgba(10, 15, 28, 0.82)"; ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = "#c9fff1"; ctx.font = "bold 42px Inter, sans-serif"; ctx.textAlign = "center";
        ctx.fillText("Prototype complete", canvas.width / 2, canvas.height / 2);
      }
    }

    function loop() {
      if (!won) update();
      draw();
      requestAnimationFrame(loop);
    }
    addEventListener("keydown", (event) => keys.add(event.key));
    addEventListener("keyup", (event) => keys.delete(event.key));
    document.getElementById("restart").addEventListener("click", reset);
    reset();
    loop();
  </script>
</body>
</html>
`)
}
