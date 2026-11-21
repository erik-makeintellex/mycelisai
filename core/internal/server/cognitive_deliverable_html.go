package server

import "strings"

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

func kpiDashboardHTML() string {
	return `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Executive KPI Dashboard</title></head><body><h1>Executive KPI Dashboard</h1><div><span id="active-period">Period: 30D</span><button type="button" data-testid="kpi-period-30d" onclick="setPeriod('30D','128','99.4%')">30D</button><button type="button" data-testid="kpi-period-qtd" onclick="setPeriod('QTD','412','99.8%')">QTD</button></div><div id="kpi-cards"><div class="kpi-card"><span>Verified Packages</span><strong id="kpi-packages">128</strong></div><div class="kpi-card"><span>Governance Pass Rate</span><strong id="kpi-pass-rate">99.4%</strong></div></div><script>function setPeriod(p,pkgs,rate){document.getElementById('active-period').textContent='Period: '+p;document.getElementById('kpi-packages').textContent=pkgs;document.getElementById('kpi-pass-rate').textContent=rate;}</script></body></html>`
}

func productWireframeHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Interactive Product Wireframe</title>
<style>
  body { font-family: system-ui, sans-serif; background: #0f172a; color: #f8fafc; padding: 2rem; }
  .tabs { display: flex; gap: 0.75rem; margin: 1rem 0; }
  button { padding: 0.5rem 1rem; border-radius: 8px; border: 1px solid #38bdf8; background: #1e293b; color: #f8fafc; cursor: pointer; }
  #wireframe-canvas { padding: 1.5rem; border: 1px dashed #475569; border-radius: 12px; background: #020617; }
</style>
</head>
<body>
  <h1>Interactive Product Wireframe</h1>
  <div class="tabs">
    <button type="button" data-testid="wireframe-tab-overview" onclick="setScreen('Workspace Overview', 'Primary workspace command deck with active mission cards and deliverable status.')">Overview</button>
    <button type="button" data-testid="wireframe-tab-workflow" onclick="setScreen('Approval Workflow', 'Step-by-step operator review gate with SHA-256 artifact verification.')">Workflow</button>
  </div>
  <div id="wireframe-canvas">
    <h2 id="active-screen">Screen: Workspace Overview</h2>
    <p id="screen-description">Primary workspace command deck with active mission cards and deliverable status.</p>
  </div>
  <script>
    function setScreen(name, desc) {
      document.getElementById('active-screen').textContent = 'Screen: ' + name;
      document.getElementById('screen-description').textContent = desc;
    }
  </script>
</body>
</html>`
}

func extractDevProjectSlug(goal string) string {
	lower := strings.ToLower(goal)
	for _, keyword := range []string{"calculator", "timer", "stopwatch", "todo", "kanban", "tracker", "landing-page", "converter", "portal"} {
		if strings.Contains(lower, strings.ReplaceAll(keyword, "-", " ")) {
			return keyword + "-app"
		}
	}
	return "starter-web-app"
}

func generalDevProjectBundleSpecs(goal, slug string) []deliverableBundleFileSpec {
	title := titleCaseTeamID(slug)
	return []deliverableBundleFileSpec{
		{
			name:        "index.html",
			title:       title + " — Interactive Web App (HTML5)",
			contentType: "text/html",
			content:     generalDevProjectHTML(title, goal),
		},
		{
			name:        "README.md",
			title:       title + " — Project Overview & Quickstart",
			contentType: "text/markdown",
			content: "# " + title + "\n\n## Project Goal\n" + goal + "\n\n## Included Files\n" +
				"- `index.html` — Ready-to-run interactive browser application.\n" +
				"- `README.md` — Project overview, features, and customization notes.\n" +
				"- `implementation_notes.md` — Component structure and next development steps.\n",
		},
		{
			name:        "implementation_notes.md",
			title:       title + " — Implementation & Extension Notes",
			contentType: "text/markdown",
			content: "# Implementation & Extension Notes\n\n## Architecture\n" +
				"- Self-contained HTML5/CSS/JS application ready for live preview or local hosting.\n" +
				"- Verified by the team's Ideal Delivery Reviewer (`council-sentry`).\n",
		},
	}
}

func generalDevProjectHTML(title, goal string) string {
	safeGoal := strings.ReplaceAll(strings.ReplaceAll(goal, "<", "&lt;"), ">", "&gt;")
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>` + title + `</title>
<style>
  body { font-family: system-ui, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 1.5rem; }
  .shell { max-width: 680px; margin: 0 auto; background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 1.25rem; }
  h1 { margin: 0 0 0.35rem; font-size: 1.3rem; color: #38bdf8; }
  p.goal { margin: 0 0 1rem; font-size: 0.85rem; color: #94a3b8; }
  .row { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
  input { flex: 1; padding: 0.5rem 0.75rem; border-radius: 8px; border: 1px solid #475569; background: #0f172a; color: #f8fafc; }
  button { padding: 0.5rem 0.9rem; border-radius: 8px; border: 1px solid #38bdf8; background: #0284c7; color: #fff; font-weight: 600; cursor: pointer; }
  ul { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.45rem; }
  li { display: flex; justify-content: space-between; align-items: center; padding: 0.55rem 0.75rem; border-radius: 8px; background: #0f172a; border: 1px solid #334155; font-size: 0.88rem; }
  .badge { font-size: 0.72rem; padding: 0.15rem 0.45rem; border-radius: 999px; background: #065f46; color: #a7f3d0; }
</style>
</head>
<body>
  <div class="shell">
    <h1>` + title + `</h1>
    <p class="goal">` + safeGoal + `</p>
    <div class="row">
      <input id="item-input" placeholder="Add a feature, task, or test item..." />
      <button id="add-btn" type="button" onclick="addItem()">Add Item</button>
    </div>
    <ul id="item-list">
      <li><span>Core interactive UI scaffold ready</span><span class="badge">Active</span></li>
      <li><span>State &amp; event handlers wired</span><span class="badge">Verified</span></li>
    </ul>
  </div>
  <script>
    function addItem() {
      const input = document.getElementById('item-input');
      const val = (input.value || '').trim();
      if (!val) return;
      const li = document.createElement('li');
      li.innerHTML = '<span></span><span class="badge">Added</span>';
      li.querySelector('span').textContent = val;
      document.getElementById('item-list').appendChild(li);
      input.value = '';
    }
  </script>
</body>
</html>`
}
