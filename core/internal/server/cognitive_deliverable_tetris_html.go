package server

func streetFighterTetrisHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Street Fighter II Arcade Tetris</title>
<style>
  :root { --accent: #ef4444; --bg: #090d16; --panel: #111827; }
  * { box-sizing: border-box; }
  body { margin: 0; font-family: system-ui, sans-serif; background: radial-gradient(circle at top, #1e1b4b, var(--bg) 70%); color: #f8fafc; display: flex; flex-direction: column; align-items: center; padding: 1rem; }
  h1 { margin: 0.25rem 0; font-size: 1.35rem; letter-spacing: 0.08em; text-transform: uppercase; color: #fde047; text-shadow: 0 0 10px var(--accent); }
  .subtitle { font-size: 0.8rem; color: #cbd5e1; margin-bottom: 0.6rem; }
  #fighter-select { display: flex; gap: 0.5rem; flex-wrap: wrap; justify-content: center; margin-bottom: 0.75rem; }
  #fighter-select button { padding: 0.4rem 0.75rem; border-radius: 6px; border: 1px solid #475569; background: #1e293b; color: #f8fafc; font-weight: 700; cursor: pointer; font-size: 0.8rem; }
  #fighter-select button.active { border-color: var(--accent); box-shadow: 0 0 10px var(--accent); background: #0f172a; color: #fde047; }
  .cabinet { display: flex; gap: 1rem; align-items: flex-start; flex-wrap: wrap; justify-content: center; background: var(--panel); padding: 0.9rem; border-radius: 12px; border: 2px solid var(--accent); box-shadow: 0 0 24px rgba(0,0,0,0.65); }
  #tetris-canvas { background: #020617; border: 2px solid #334155; border-radius: 6px; display: block; }
  .side-panel { display: flex; flex-direction: column; gap: 0.55rem; min-width: 185px; }
  .stat-card { background: #020617; border: 1px solid #334155; border-radius: 8px; padding: 0.5rem 0.65rem; font-size: 0.82rem; }
  .stat-card strong { color: #fde047; font-size: 0.95rem; }
  #combo-callout { min-height: 2.2rem; display: flex; align-items: center; justify-content: center; text-align: center; padding: 0.4rem; border-radius: 8px; background: #1e1b4b; border: 1px solid var(--accent); color: #fef08a; font-weight: 800; font-size: 0.78rem; letter-spacing: 0.04em; }
  .controls { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.4rem; margin-top: 0.25rem; }
  .controls button { padding: 0.45rem 0.5rem; border-radius: 6px; border: 1px solid #475569; background: #1e293b; color: #f8fafc; font-weight: 700; cursor: pointer; font-size: 0.76rem; }
  .controls button:hover { border-color: var(--accent); }
</style>
</head>
<body>
  <h1>Street Fighter II Arcade Tetris</h1>
  <div class="subtitle">Select your World Warrior, clear lines, and unleash special combos!</div>
  <div id="fighter-select" role="group" aria-label="Fighter Select">
    <button type="button" data-testid="fighter-ryu" class="active" onclick="selectFighter('ryu')">Ryu</button>
    <button type="button" data-testid="fighter-chunli" onclick="selectFighter('chunli')">Chun-Li</button>
    <button type="button" data-testid="fighter-ken" onclick="selectFighter('ken')">Ken</button>
    <button type="button" data-testid="fighter-guile" onclick="selectFighter('guile')">Guile</button>
  </div>
  <div class="cabinet">
    <canvas id="tetris-canvas" width="240" height="440"></canvas>
    <div class="side-panel">
      <div class="stat-card"><div>Challenger</div><strong id="active-fighter-name">Ryu</strong><div id="active-fighter-stage" style="font-size:0.75rem;color:#94a3b8;">Suzaku Castle Crimson</div></div>
      <div class="stat-card"><div>Special Move</div><strong id="active-fighter-move">HADOUKEN!</strong></div>
      <div class="stat-card"><div>Score: <strong id="score-val">0</strong></div><div>Lines: <strong id="lines-val">0</strong> | Level: <strong id="level-val">1</strong></div></div>
      <div id="combo-callout">READY — FIGHT!</div>
      <div class="controls">
        <button id="btn-left" type="button">◀ Left</button>
        <button id="btn-rotate" type="button">⟳ Rotate</button>
        <button id="btn-right" type="button">Right ▶</button>
        <button id="btn-drop" type="button">▼ Drop</button>
        <button id="btn-hard-drop" type="button">⚡ Slam</button>
        <button id="btn-restart" type="button">↺ Reset</button>
      </div>
    </div>
  </div>
  <script>
    const FIGHTERS = {
      ryu: { name: 'Ryu', stage: 'Suzaku Castle Crimson', move: 'HADOUKEN!', accent: '#ef4444', palette: ['#ef4444','#f97316','#eab308','#38bdf8','#a855f7','#22c55e','#f43f5e'] },
      chunli: { name: 'Chun-Li', stage: 'Taiping Road Cobalt', move: 'LIGHTNING KICK!', accent: '#38bdf8', palette: ['#38bdf8','#06b6d4','#818cf8','#f472b6','#34d399','#facc15','#60a5fa'] },
      ken: { name: 'Ken', stage: 'Battle Harbor Flame', move: 'SHORYUKEN!', accent: '#f59e0b', palette: ['#f59e0b','#ef4444','#fde047','#fb923c','#f43f5e','#38bdf8','#a3e635'] },
      guile: { name: 'Guile', stage: 'Air Force Base Emerald', move: 'SONIC BOOM!', accent: '#10b981', palette: ['#10b981','#eab308','#22c55e','#14b8a6','#84cc16','#38bdf8','#f59e0b'] }
    };
    const SHAPES = [
      [[1,1],[1,1]], [[1,1,1,1]], [[0,1,0],[1,1,1]], [[0,1,1],[1,1,0]],
      [[1,1,0],[0,1,1]], [[1,0,0],[1,1,1]], [[0,0,1],[1,1,1]]
    ];
    const COLS = 10, ROWS = 20, BLOCK = 22;
    const canvas = document.getElementById('tetris-canvas'), ctx = canvas.getContext('2d');
    let currentFighter = FIGHTERS.ryu, board = [], piece = null, bagIdx = 0, score = 0, lines = 0, level = 1;
    function initBoard() {
      board = Array.from({ length: ROWS }, () => Array(COLS).fill(0));
      for (let r = ROWS - 2; r < ROWS; r++) {
        for (let c = 0; c < COLS; c++) if (c < 4 || c > 5) board[r][c] = (c % 6) + 1;
      }
    }
    function spawnPiece() {
      const idx = bagIdx % SHAPES.length; bagIdx++;
      const shape = SHAPES[idx].map(r => r.slice());
      piece = { shape, x: 4, y: 0, colorIdx: idx };
      if (collides(piece.shape, piece.x, piece.y)) { initBoard(); }
    }
    function collides(shape, px, py) {
      for (let r = 0; r < shape.length; r++) {
        for (let c = 0; c < shape[r].length; c++) {
          if (!shape[r][c]) continue;
          const nx = px + c, ny = py + r;
          if (nx < 0 || nx >= COLS || ny >= ROWS || (ny >= 0 && board[ny][nx])) return true;
        }
      }
      return false;
    }
    function mergeAndClear() {
      piece.shape.forEach((row, r) => row.forEach((v, c) => {
        if (v && piece.y + r >= 0) board[piece.y + r][piece.x + c] = piece.colorIdx + 1;
      }));
      let cleared = 0;
      for (let r = ROWS - 1; r >= 0; r--) {
        if (board[r].every(v => v !== 0)) { board.splice(r, 1); board.unshift(Array(COLS).fill(0)); cleared++; r++; }
      }
      if (cleared > 0) {
        const pts = cleared * 100 * level;
        score += pts; lines += cleared; level = 1 + Math.floor(lines / 5);
        document.getElementById('combo-callout').textContent = currentFighter.name.toUpperCase() + ' UNLEASHES ' + currentFighter.move + ' +' + pts;
      }
      updateHud(); spawnPiece(); draw();
    }
    function move(dx) { if (!collides(piece.shape, piece.x + dx, piece.y)) { piece.x += dx; draw(); } }
    function rotate() {
      const rot = piece.shape[0].map((_, c) => piece.shape.map(r => r[c]).reverse());
      if (!collides(rot, piece.x, piece.y)) { piece.shape = rot; draw(); }
    }
    function softDrop() {
      if (!collides(piece.shape, piece.x, piece.y + 1)) { piece.y++; score += 1; updateHud(); draw(); }
      else mergeAndClear();
    }
    function hardDrop() {
      while (!collides(piece.shape, piece.x, piece.y + 1)) { piece.y++; score += 2; }
      mergeAndClear();
    }
    function updateHud() {
      document.getElementById('score-val').textContent = String(score);
      document.getElementById('lines-val').textContent = String(lines);
      document.getElementById('level-val').textContent = String(level);
    }
    function draw() {
      ctx.fillStyle = '#020617'; ctx.fillRect(0, 0, canvas.width, canvas.height);
      const ox = 10, oy = 0;
      for (let r = 0; r < ROWS; r++) for (let c = 0; c < COLS; c++) {
        ctx.strokeStyle = '#0f172a'; ctx.strokeRect(ox + c * BLOCK, oy + r * BLOCK, BLOCK, BLOCK);
        if (board[r][c]) {
          ctx.fillStyle = currentFighter.palette[(board[r][c] - 1) % currentFighter.palette.length];
          ctx.fillRect(ox + c * BLOCK + 1, oy + r * BLOCK + 1, BLOCK - 2, BLOCK - 2);
        }
      }
      if (piece) piece.shape.forEach((row, r) => row.forEach((v, c) => {
        if (v) {
          ctx.fillStyle = currentFighter.palette[piece.colorIdx % currentFighter.palette.length];
          ctx.fillRect(ox + (piece.x + c) * BLOCK + 1, oy + (piece.y + r) * BLOCK + 1, BLOCK - 2, BLOCK - 2);
        }
      }));
    }
    window.selectFighter = function(key) {
      currentFighter = FIGHTERS[key] || FIGHTERS.ryu;
      document.documentElement.style.setProperty('--accent', currentFighter.accent);
      document.getElementById('active-fighter-name').textContent = currentFighter.name;
      document.getElementById('active-fighter-stage').textContent = currentFighter.stage;
      document.getElementById('active-fighter-move').textContent = currentFighter.move;
      document.getElementById('combo-callout').textContent = currentFighter.name.toUpperCase() + ' READY — ' + currentFighter.move;
      document.querySelectorAll('#fighter-select button').forEach(b => b.classList.toggle('active', b.getAttribute('data-testid') === 'fighter-' + key));
      draw();
    };
    function restartGame() { score = 0; lines = 0; level = 1; bagIdx = 0; initBoard(); spawnPiece(); updateHud(); draw(); }
    document.getElementById('btn-left').addEventListener('click', () => move(-1));
    document.getElementById('btn-right').addEventListener('click', () => move(1));
    document.getElementById('btn-rotate').addEventListener('click', rotate);
    document.getElementById('btn-drop').addEventListener('click', softDrop);
    document.getElementById('btn-hard-drop').addEventListener('click', hardDrop);
    document.getElementById('btn-restart').addEventListener('click', restartGame);
    window.addEventListener('keydown', e => {
      if (e.key === 'ArrowLeft') move(-1);
      else if (e.key === 'ArrowRight') move(1);
      else if (e.key === 'ArrowUp') rotate();
      else if (e.key === 'ArrowDown') softDrop();
      else if (e.key === ' ') { e.preventDefault(); hardDrop(); }
    });
    restartGame();
    setInterval(softDrop, 850);
  </script>
</body>
</html>`
}

func streetFighterTetrisBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{
			name: "index.html", title: "Playable Street Fighter II Tetris (HTML5)", contentType: "text/html",
			content: streetFighterTetrisHTML(),
		},
		{
			name: "fighter_roster_and_theme.md", title: "Fighter Roster & Theme Guide", contentType: "text/markdown",
			content: "# Fighter Roster & Theme Guide\n\n## Request Context\n" + goal + "\n\n## Playable World Warriors\n" +
				"- **Ryu (`Suzaku Castle Crimson`):** Accent `#ef4444` — unleashes `HADOUKEN!` on line clears.\n" +
				"- **Chun-Li (`Taiping Road Cobalt`):** Accent `#38bdf8` — unleashes `LIGHTNING KICK!` on line clears.\n" +
				"- **Ken (`Battle Harbor Flame`):** Accent `#f59e0b` — unleashes `SHORYUKEN!` on line clears.\n" +
				"- **Guile (`Air Force Base Emerald`):** Accent `#10b981` — unleashes `SONIC BOOM!` on line clears.\n",
		},
		{
			name: "controls_and_scoring.md", title: "Arcade Controls & Scoring Rules", contentType: "text/markdown",
			content: "# Arcade Controls & Scoring Rules\n\n## Controls\n" +
				"- **Keyboard:** `ArrowLeft` / `ArrowRight` move, `ArrowUp` rotates, `ArrowDown` soft drops, `Space` slams.\n" +
				"- **Arcade Deck Buttons:** `#btn-left`, `#btn-rotate`, `#btn-right`, `#btn-drop`, `#btn-hard-drop`, `#btn-restart`.\n\n" +
				"## Combo Scoring\n- Clearing 1+ lines triggers the active fighter's special move banner (`#combo-callout`) and awards `100 * lines * level` points.\n",
		},
	}
}
