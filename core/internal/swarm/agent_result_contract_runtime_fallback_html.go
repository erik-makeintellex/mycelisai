package swarm

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func runtimeFallbackProjectPackageHTML(title string, requirement *teamResultRequirement) string {
	escapedTitle := html.EscapeString(title)
	actionText := runtimeFallbackActionText(requirement)
	keyHandler := runtimeFallbackKeyHandler(requirement)
	payloadJSON, _ := json.Marshal(map[string]any{
		"runtime_owned": true,
		"recovered":     true,
		"purpose":       "approved project package fallback",
	})
	expectedKey := ""
	if requirement != nil && requirement.OutputValidation != nil && requirement.OutputValidation.Probe != nil {
		expectedKey = requirement.OutputValidation.Probe.Action.Key
	}
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + escapedTitle + `</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, sans-serif; background: #10151d; color: #eef5ff; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: radial-gradient(circle at top left, #1e3b38, #10151d 52%); }
    main { width: min(760px, calc(100vw - 32px)); border: 1px solid #38535f; border-radius: 18px; padding: 28px; background: rgba(17, 24, 34, .92); box-shadow: 0 24px 70px rgba(0,0,0,.35); }
    h1 { margin: 0 0 10px; font-size: clamp(1.8rem, 4vw, 3rem); }
    p { color: #bfd0df; line-height: 1.55; }
    canvas { display: block; width: 100%; max-width: 640px; height: auto; border: 1px solid #31535b; border-radius: 14px; background: #0a1018; }
    .surface { margin: 22px 0; padding: 18px; border-radius: 14px; background: #17242c; border: 1px solid #31535b; font-weight: 700; color: #90f0dd; }
    button { appearance: none; border: 0; border-radius: 999px; padding: 14px 22px; background: #7ee4cf; color: #06231f; font-weight: 800; cursor: pointer; }
    button + button { margin-left: 10px; background: #263543; color: #dbe8f3; }
  </style>
</head>
<body>
  <main>
    <h1>` + escapedTitle + `</h1>
    <p>` + actionText + `</p>
    <canvas id="game" width="640" height="360" aria-label="Playable recovered package"></canvas>
    <div class="surface" data-mycelis-validation-surface id="validationSurface">Ready for interaction. Score: 0</div>
    <button data-mycelis-primary-action id="primaryAction" type="button">Play</button>
    <button id="resetAction" type="button">Restart</button>
  </main>
  <script type="application/json" id="mycelis-package-proof">` + html.EscapeString(string(payloadJSON)) + `</script>
  <script>
    const canvas = document.getElementById('game');
    const ctx = canvas.getContext('2d');
    const surface = document.querySelector('[data-mycelis-validation-surface]');
    const primary = document.querySelector('[data-mycelis-primary-action]');
    const reset = document.getElementById('resetAction');
    const expectedKey = ` + fmt.Sprintf("%q", expectedKey) + `;
    let score = 0;
    let playerX = 72;
    let playerY = 260;
    let vx = 0;
    function draw() {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      const gradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height);
      gradient.addColorStop(0, '#102934');
      gradient.addColorStop(1, '#171f2b');
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.fillStyle = '#304455';
      ctx.fillRect(0, 304, canvas.width, 56);
      ctx.fillStyle = '#ffd166';
      ctx.fillRect(500, 264, 32, 40);
      ctx.fillStyle = '#7ee4cf';
      ctx.fillRect(playerX, playerY, 34, 44);
      ctx.fillStyle = '#eef5ff';
      ctx.font = '18px sans-serif';
      ctx.fillText('Hold ArrowRight or press Play', 24, 36);
      playerX = Math.max(24, Math.min(584, playerX + vx));
      requestAnimationFrame(draw);
    }
    function advance(reason) {
      score += 1;
      playerX = Math.min(584, playerX + 18);
      surface.textContent = reason + ' Score: ' + score;
      surface.dataset.state = 'changed-' + score;
    }
    primary.addEventListener('click', () => advance('Primary action completed.'));
    primary.addEventListener('pointerdown', () => surface.dataset.pointer = 'ready');
    window.addEventListener('keydown', (event) => {
      if (event.key === 'ArrowRight') {
        vx = 4;
        advance('ArrowRight movement accepted.');
      }
    });
    window.addEventListener('keyup', (event) => {
      if (event.key === 'ArrowRight') {
        vx = 0;
      }
    });
    reset.addEventListener('click', () => {
      score = 0;
      playerX = 72;
      vx = 0;
      surface.textContent = 'Ready for interaction. Score: 0';
      surface.dataset.state = 'reset';
    });` + keyHandler + `
    draw();
  </script>
</body>
</html>
`
}

func runtimeFallbackActionText(requirement *teamResultRequirement) string {
	actionText := "Use the primary action below to confirm the package changes state."
	if requirement != nil && len(requirement.AcceptanceCriteria) > 0 {
		actionText = html.EscapeString("Contract focus: " + strings.Join(requirement.AcceptanceCriteria, "; "))
	}
	if key := runtimeFallbackExpectedKey(requirement); key != "" {
		actionText += " Keyboard action: " + html.EscapeString(key) + "."
	}
	return actionText
}

func runtimeFallbackKeyHandler(requirement *teamResultRequirement) string {
	if runtimeFallbackExpectedKey(requirement) == "" {
		return ""
	}
	return `
window.addEventListener('keydown', (event) => {
  if (!expectedKey || event.key.toLowerCase() === expectedKey.toLowerCase()) {
    advance('Key action accepted.');
  }
});`
}

func runtimeFallbackExpectedKey(requirement *teamResultRequirement) string {
	if requirement == nil || requirement.OutputValidation == nil || requirement.OutputValidation.Probe == nil {
		return ""
	}
	action := requirement.OutputValidation.Probe.Action
	if action.Kind != protocol.OutputValidationActionKeyPress && action.Kind != protocol.OutputValidationActionKeyHold {
		return ""
	}
	if strings.TrimSpace(action.Key) == "" {
		return "any key"
	}
	return action.Key
}
