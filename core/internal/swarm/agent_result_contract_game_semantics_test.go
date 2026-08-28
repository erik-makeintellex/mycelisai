package swarm

import (
	"strings"
	"testing"
)

var functionalGameCriteria = []string{
	"playable controls move the player and change the visible game surface",
	"attack changes enemy, hazard, or score state",
}

func TestGameSemanticValidationRejectsLabelAndCanvasTextFacade(t *testing.T) {
	content := `<!doctype html>
<p data-mycelis-game-instructions>ArrowRight moves. Attack enemies and restart.</p>
<button data-mycelis-validation-action="attack">Attack</button>
<canvas id="game"></canvas><output id="score">0</output>
<script>
const game = document.getElementById('game');
let labelState = 'ready';
addEventListener('keydown', event => { if (event.key === 'ArrowRight') labelState = 'moved'; game.textContent = labelState; });
document.querySelector('[data-mycelis-validation-action="attack"]').onclick = () => score.textContent = '10';
</script>`
	issues := strings.Join(semanticAcceptanceEntrypointIssues(functionalGameCriteria, content), ";")
	for _, expected := range []string{
		"text-label facade", "no inspectable 2d render", "no active canvas render", "no governed player", "movement controls do not mutate", "attack action does not mutate",
	} {
		if !strings.Contains(issues, expected) {
			t.Fatalf("issues = %q, want %q", issues, expected)
		}
	}
}

func TestGameSemanticValidationRejectsUnrelatedStatusMutationWithDecorativeLoop(t *testing.T) {
	content := `<!doctype html>
<p data-mycelis-game-instructions>ArrowRight moves. Attack enemies and restart.</p>
<button data-mycelis-validation-action="attack">Attack</button>
<canvas id="game"></canvas><output id="score">0</output><p id="status">Ready</p>
<script>
const canvas = document.getElementById('game'); const ctx = canvas.getContext('2d');
const player = {x: 10, y: 10}; const enemies = [{health: 1}];
addEventListener('keydown', event => { if (event.key === 'ArrowRight') status.textContent = 'Moved'; });
document.querySelector('[data-mycelis-validation-action="attack"]').onclick = () => status.textContent = 'Attacked';
function gameLoop(){ ctx.clearRect(0,0,100,100); ctx.fillRect(player.x, player.y, 10, 10); requestAnimationFrame(gameLoop); } gameLoop();
</script>`
	issues := strings.Join(semanticAcceptanceEntrypointIssues(functionalGameCriteria, content), ";")
	for _, expected := range []string{"movement controls do not mutate", "attack action does not mutate"} {
		if !strings.Contains(issues, expected) {
			t.Fatalf("issues = %q, want %q", issues, expected)
		}
	}
}

func TestGameSemanticValidationAcceptsStateDrivenCanvasLoop(t *testing.T) {
	content := `<!doctype html>
<p data-mycelis-game-instructions>ArrowRight moves. Attack enemies and restart.</p>
<button data-mycelis-validation-action="attack">Attack</button>
<canvas id="game"></canvas><output id="score">0</output>
<script>
const canvas = document.getElementById('game'); const ctx = canvas.getContext('2d');
const player = {x: 10, y: 10, vx: 0}; const enemies = [{health: 1}]; let score = 0;
addEventListener('keydown', event => { if (event.key === 'ArrowRight') player.vx = 2; });
function attack(){ enemies[0].health -= 1; score += 10; }
document.querySelector('[data-mycelis-validation-action="attack"]').onclick = attack;
function gameLoop(){ player.x += player.vx; ctx.clearRect(0,0,100,100); ctx.fillRect(player.x, player.y, 10, 10); requestAnimationFrame(gameLoop); } gameLoop();
</script>`
	if issues := semanticAcceptanceEntrypointIssues(functionalGameCriteria, content); len(issues) != 0 {
		t.Fatalf("functional state-driven game was rejected: %v", issues)
	}
}

func TestGameSemanticValidationDoesNotChangeGenericAppPosture(t *testing.T) {
	criteria := []string{"primary interaction changes the application state"}
	content := `<canvas id="game"></canvas><button onclick="game.textContent='Changed'">Run</button>`
	if issues := semanticAcceptanceEntrypointIssues(criteria, content); len(issues) != 0 {
		t.Fatalf("generic app received game-only semantic issues: %v", issues)
	}
}

func TestGameSemanticValidationRejectsRequiredActionsThatOnlyChangeLabels(t *testing.T) {
	criteria := append([]string{}, functionalGameCriteria...)
	criteria = append(criteria,
		"hazard contact changes health state",
		"key pickup changes key and score state",
		"team play-tests the documented winning route through the locked door objective to win state",
		"fail state can transition through restart to the initial objective",
	)
	content := `<!doctype html><p data-mycelis-game-instructions>ArrowRight moves. Attack, collect the key, win, fail, and restart.</p>
<canvas id="game"></canvas><button data-mycelis-validation-action="attack">Attack</button>
<button data-mycelis-validation-action="hazard">Hazard</button><button data-mycelis-validation-action="key">Key</button>
<button data-mycelis-validation-action="win">Win</button><button data-mycelis-validation-action="fail">Fail</button>
<button data-mycelis-validation-action="restart">Restart</button><span id="score">0</span><span id="health">4</span>
<span id="keyState">No</span><span id="goalState">Find key</span><script>
const canvas=document.getElementById('game'),ctx=canvas.getContext('2d'),player={x:0,y:0,vx:0},enemies=[];
addEventListener('keydown',e=>{if(e.key==='ArrowRight')player.vx=2});
attack.onclick=()=>score.textContent='10'; hazard.onclick=()=>health.textContent='3'; key.onclick=()=>keyState.textContent='Yes';
win.onclick=()=>goalState.textContent='Won'; fail.onclick=()=>goalState.textContent='Failed'; restart.onclick=()=>goalState.textContent='Find key';
function loop(){player.x+=player.vx;ctx.fillRect(player.x,player.y,5,5);requestAnimationFrame(loop)}loop();</script>`
	issues := strings.Join(semanticAcceptanceEntrypointIssues(criteria, content), ";")
	for _, expected := range []string{"attack action", "hazard action", "key action", "win action", "fail and restart actions"} {
		if !strings.Contains(issues, expected) {
			t.Fatalf("issues = %q, want %q", issues, expected)
		}
	}
}
