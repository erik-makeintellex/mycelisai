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
			"validation":         "Retained as a self-contained browser adventure with movement, collision, hazards, enemies, key, door, win/fail states, and restart.",
		},
	}, true
}

func firstTeamGamePackageFiles(request string) []string {
	files := []string{"index.html", "README.md", "PROOF.md", "project-package.json"}
	if projectPackageTextRequestsReadme(request) {
		files = append(files, "README.md")
	}
	return uniqueOrderedTools(files)
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
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, sans-serif; }
    * { box-sizing: border-box; }
    body {
      margin: 0; min-height: 100vh; display: grid; place-items: center;
      background: radial-gradient(circle at 50% 0%, #21334b 0, #0a101c 54%, #05070c 100%);
      color: #eef4ff;
    }
    main { width: min(980px, calc(100vw - 28px)); display: grid; gap: 14px; padding: 18px 0; }
    header { display: flex; justify-content: space-between; gap: 12px; align-items: end; }
    h1 { margin: 0; font-size: clamp(22px, 4vw, 32px); letter-spacing: 0; }
    p { margin: 5px 0 0; color: #aab8d4; line-height: 1.45; max-width: 720px; }
    .hud { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; min-width: 260px; }
    .pill { border: 1px solid #344661; border-radius: 7px; padding: 8px 10px; background: #111b2c; font-weight: 800; }
    canvas { width: 100%; aspect-ratio: 16 / 9; border: 1px solid #344661; border-radius: 8px; background: #101827; outline: none; }
    .controls { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; color: #aab8d4; }
    button { border: 1px solid #71e0c3; border-radius: 6px; background: #153a33; color: #d9fff6; padding: 10px 14px; font-weight: 800; cursor: pointer; }
    button:hover { background: #1b4a40; }
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
        <div class="pill">Health <span id="health">3</span></div>
        <div class="pill">Key <span id="keyState">No</span></div>
        <div class="pill">Goal <span id="goalState">Find relic</span></div>
      </div>
    </header>
    <canvas id="game" width="960" height="540" tabindex="0" aria-label="Playable top-down adventure"></canvas>
    <div class="controls">
      <button id="restart">Restart</button>
      <span>Move with arrows or WASD. Take the key, avoid brambles and sentries, unlock the north gate, and claim the relic. Press R to restart.</span>
    </div>
  </main>
  <script>
    const canvas = document.getElementById("game");
    const ctx = canvas.getContext("2d");
    const healthEl = document.getElementById("health");
    const keyEl = document.getElementById("keyState");
    const goalEl = document.getElementById("goalState");
    const keys = new Set();
    const TILE = 48;
    const player = { x: 72, y: 432, w: 28, h: 32, speed: 3.1, invuln: 0 };
    const walls = [
      { x: 0, y: 0, w: 960, h: 48 }, { x: 0, y: 492, w: 960, h: 48 },
      { x: 0, y: 0, w: 48, h: 540 }, { x: 912, y: 0, w: 48, h: 540 },
      { x: 144, y: 144, w: 96, h: 48 }, { x: 336, y: 96, w: 48, h: 192 },
      { x: 528, y: 336, w: 192, h: 48 }, { x: 720, y: 144, w: 48, h: 192 },
      { x: 144, y: 336, w: 48, h: 96 },
      { x: 384, y: 48, w: 48, h: 144 }, { x: 528, y: 48, w: 48, h: 144 },
      { x: 384, y: 192, w: 60, h: 48 }, { x: 516, y: 192, w: 60, h: 48 }
    ];
    const hazards = [
      { x: 252, y: 348, w: 132, h: 36 }, { x: 576, y: 120, w: 96, h: 36 }
    ];
    const keyItemStart = { x: 828, y: 420, w: 24, h: 24 };
    const heartStart = { x: 94, y: 96, w: 24, h: 24 };
    const door = { x: 444, y: 192, w: 72, h: 48 };
    const relic = { x: 462, y: 84, w: 36, h: 36 };
    let enemies = [];
    let health = 3;
    let hasKey = false;
    let keyItem = { ...keyItemStart };
    let heart = { ...heartStart };
    let heartTaken = false;
    let state = "playing";
    let tick = 0;

    function reset() {
      Object.assign(player, { x: 72, y: 432, invuln: 0 });
      health = 3; hasKey = false; keyItem = { ...keyItemStart }; heart = { ...heartStart };
      heartTaken = false; state = "playing"; tick = 0;
      enemies = [
        { x: 250, y: 245, w: 30, h: 30, axis: "x", base: 250, range: 95, speed: 0.033, phase: 0 },
        { x: 626, y: 250, w: 30, h: 30, axis: "y", base: 250, range: 115, speed: 0.027, phase: 1.4 },
        { x: 776, y: 384, w: 30, h: 30, axis: "x", base: 776, range: 76, speed: 0.041, phase: 2.2 }
      ];
      syncHud();
      canvas.focus();
    }

    function syncHud() {
      healthEl.textContent = health;
      keyEl.textContent = hasKey ? "Yes" : "No";
      goalEl.textContent = state === "won" ? "Won" : state === "failed" ? "Failed" : hasKey ? "Open gate" : "Find key";
    }

    function hit(a, b) {
      return a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y;
    }

    function blockers() {
      return hasKey ? walls : walls.concat(door);
    }

    function movePlayer(dx, dy) {
      player.x += dx;
      for (const wall of blockers()) if (hit(player, wall)) player.x -= dx;
      player.y += dy;
      for (const wall of blockers()) if (hit(player, wall)) player.y -= dy;
    }

    function damage(amount) {
      if (player.invuln > 0 || state !== "playing") return;
      health = Math.max(0, health - amount);
      player.invuln = 54;
      if (health === 0) state = "failed";
      syncHud();
    }

    function update() {
      tick += 1;
      if (player.invuln > 0) player.invuln -= 1;
      if (state !== "playing") return;
      const dx = (keys.has("ArrowRight") || keys.has("d") ? player.speed : 0) - (keys.has("ArrowLeft") || keys.has("a") ? player.speed : 0);
      const dy = (keys.has("ArrowDown") || keys.has("s") ? player.speed : 0) - (keys.has("ArrowUp") || keys.has("w") ? player.speed : 0);
      movePlayer(dx, dy);
      for (const enemy of enemies) {
        const offset = Math.sin(tick * enemy.speed + enemy.phase) * enemy.range;
        if (enemy.axis === "x") enemy.x = enemy.base + offset;
        else enemy.y = enemy.base + offset;
        if (hit(player, enemy)) damage(1);
      }
      for (const hazard of hazards) if (hit(player, hazard)) damage(1);
      if (!hasKey && hit(player, keyItem)) { hasKey = true; syncHud(); }
      if (!heartTaken && hit(player, heart)) { heartTaken = true; health = Math.min(3, health + 1); syncHud(); }
      if (hasKey && hit(player, relic)) { state = "won"; syncHud(); }
      syncHud();
    }

    function rect(box, fill, stroke) {
      ctx.fillStyle = fill; ctx.fillRect(box.x, box.y, box.w, box.h);
      if (stroke) { ctx.strokeStyle = stroke; ctx.lineWidth = 2; ctx.strokeRect(box.x + 1, box.y + 1, box.w - 2, box.h - 2); }
    }

    function draw() {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      const gradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height);
      gradient.addColorStop(0, "#143922");
      gradient.addColorStop(1, "#0b1720");
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.strokeStyle = "rgba(184, 222, 175, 0.08)";
      for (let x = 0; x < canvas.width; x += TILE) { ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, canvas.height); ctx.stroke(); }
      for (let y = 0; y < canvas.height; y += TILE) { ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(canvas.width, y); ctx.stroke(); }
      walls.forEach((wall) => rect(wall, "#243447", "#415877"));
      hazards.forEach((hazard, i) => {
        rect(hazard, i % 2 ? "#783225" : "#6b2340", "#f29466");
        ctx.fillStyle = "rgba(255, 220, 122, 0.45)";
        for (let x = hazard.x + 8; x < hazard.x + hazard.w; x += 24) {
          ctx.beginPath(); ctx.arc(x, hazard.y + 18 + Math.sin((tick + x) / 12) * 4, 5, 0, Math.PI * 2); ctx.fill();
        }
      });
      if (!hasKey) {
        ctx.save(); ctx.translate(keyItem.x + 12, keyItem.y + 12); ctx.rotate(tick / 35);
        rect({ x: -10, y: -5, w: 20, h: 10 }, "#f5d76e", "#fff2aa");
        rect({ x: 4, y: -3, w: 14, h: 6 }, "#f5d76e", "#fff2aa"); ctx.restore();
      }
      if (!heartTaken) {
        ctx.fillStyle = "#ff7895"; ctx.beginPath();
        ctx.arc(heart.x + 8, heart.y + 9, 8, 0, Math.PI * 2); ctx.arc(heart.x + 16, heart.y + 9, 8, 0, Math.PI * 2);
        ctx.moveTo(heart.x + 2, heart.y + 14); ctx.lineTo(heart.x + 12, heart.y + 26); ctx.lineTo(heart.x + 22, heart.y + 14); ctx.fill();
      }
      rect(door, hasKey ? "#397b67" : "#7d6330", hasKey ? "#75f0cf" : "#ffd86b");
      rect(relic, "#8d7cff", "#d8d1ff");
      ctx.fillStyle = "#fff5b8"; ctx.beginPath(); ctx.arc(relic.x + 18, relic.y + 18, 8 + Math.sin(tick / 10) * 3, 0, Math.PI * 2); ctx.fill();
      enemies.forEach((enemy) => {
        rect(enemy, "#d95664", "#ffb0b8");
        ctx.fillStyle = "#19090d"; ctx.fillRect(enemy.x + 7, enemy.y + 9, 5, 5); ctx.fillRect(enemy.x + 18, enemy.y + 9, 5, 5);
      });
      rect(player, player.invuln % 10 > 5 ? "#d5fff5" : "#70b2ff", "#e4f3ff");
      ctx.fillStyle = "#122034"; ctx.fillRect(player.x + 7, player.y + 8, 5, 5); ctx.fillRect(player.x + 18, player.y + 8, 5, 5);
      if (state !== "playing") {
        ctx.fillStyle = "rgba(10, 15, 28, 0.82)"; ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = state === "won" ? "#c9fff1" : "#ffc0cc"; ctx.font = "bold 42px Inter, sans-serif"; ctx.textAlign = "center";
        ctx.fillText(state === "won" ? "Relic recovered" : "Expedition failed", canvas.width / 2, canvas.height / 2 - 18);
        ctx.font = "20px Inter, sans-serif"; ctx.fillText("Press R or Restart to play again", canvas.width / 2, canvas.height / 2 + 26);
      }
    }

    function loop() {
      update();
      draw();
      requestAnimationFrame(loop);
    }
    addEventListener("keydown", (event) => { if (event.key === "r" || event.key === "R") reset(); keys.add(event.key); });
    addEventListener("keyup", (event) => keys.delete(event.key));
    document.getElementById("restart").addEventListener("click", reset);
    canvas.addEventListener("click", () => canvas.focus());
    reset();
    loop();
  </script>
</body>
</html>
`)
}
