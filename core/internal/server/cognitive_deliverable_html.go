package server

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
