package outputvalidation

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runtimeValidator(t *testing.T) *PlaywrightValidator {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is not installed for the optional Playwright adapter test")
	}
	scriptPath, err := filepath.Abs(filepath.Join("..", "..", "..", "interface", "scripts", "validate-generated-output.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	modulePath := filepath.Join(filepath.Dir(filepath.Dir(scriptPath)), "node_modules", "@playwright", "test")
	if _, err := os.Stat(modulePath); err != nil {
		t.Skip("interface Playwright dependency is not installed; production must configure an available runner")
	}
	validator, err := NewPlaywrightValidator(PlaywrightConfig{ScriptPath: scriptPath, Timeout: 30 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return validator
}

func serveValidationFixtures(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/interactive", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<!doctype html><title>fixture</title><style>canvas{width:160px;height:80px}</style>
<canvas id="game" width="160" height="80"></canvas><script>
const c=document.querySelector('#game'),x=c.getContext('2d');let px=10,right=false;
function draw(){x.fillStyle='#000';x.fillRect(0,0,160,80);x.fillStyle='#0f0';x.fillRect(px,20,12,12);if(right)px+=2;requestAnimationFrame(draw)}
addEventListener('keydown',e=>{if(e.key==='ArrowRight')right=true});addEventListener('keyup',e=>{if(e.key==='ArrowRight')right=false});draw();</script>`)
	})
	mux.HandleFunc("/actions", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<!doctype html><title>actions</title><button id="click" onclick="document.querySelector('#status').textContent='changed'">Run</button>
<button id="show" onclick="document.querySelector('#hidden').style.display='block'">Show</button><button id="route" onclick="history.pushState({},'', '/changed')">Route</button>
<input id="input"><p id="status">initial</p><p id="hidden" style="display:none">visible</p>
<script>const input=document.querySelector('#input');input.addEventListener('keydown',e=>{if(e.key==='Enter')input.value='pressed'})</script>`)
	})
	mux.HandleFunc("/broken", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<!doctype html><img src="/missing.png"><script>throw new Error('fixture page error')</script>`)
	})
	mux.HandleFunc("/static", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<!doctype html><canvas id="game" width="160" height="80"></canvas><script>
const x=game.getContext('2d');x.fillStyle='#000';x.fillRect(0,0,160,80);</script>`)
	})
	return httptest.NewServer(mux)
}

func validateFixture(t *testing.T, validator *PlaywrightValidator, launchURL string, probe *Probe) Report {
	t.Helper()
	evidencePath := t.TempDir()
	report, err := validator.Validate(context.Background(), Request{
		LaunchURL: launchURL, ContentDigest: "sha256:fixture", EvidencePath: evidencePath,
		Plan: Plan{
			Kind: KindInteractiveBrowser, Required: true,
			Checks: []Check{CheckLoad, CheckNoPageErrors, CheckNoFailedLocalAsset}, Probe: probe,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range report.EvidenceRefs {
		if !strings.HasPrefix(filepath.Clean(ref.Path), filepath.Clean(evidencePath)+string(os.PathSeparator)) {
			t.Fatalf("evidence escaped retained path: %#v", ref)
		}
	}
	return report
}

func TestPlaywrightRuntimeGenericProbes(t *testing.T) {
	validator := runtimeValidator(t)
	server := serveValidationFixtures(t)
	defer server.Close()
	tests := []struct {
		name  string
		path  string
		probe Probe
	}{
		{"key hold visual change", "/interactive", Probe{
			Action:  ProbeAction{Kind: ActionKeyHold, Key: "ArrowRight", DurationMS: 400},
			Observe: ProbeObservation{Kind: ObserveVisualChange, Target: "#game"},
		}},
		{"click text change", "/actions", Probe{
			Action:  ProbeAction{Kind: ActionClick, Target: "#click"},
			Observe: ProbeObservation{Kind: ObserveTextChange, Target: "#status"},
		}},
		{"key press value change", "/actions", Probe{
			Action:  ProbeAction{Kind: ActionKeyPress, Target: "#input", Key: "Enter"},
			Observe: ProbeObservation{Kind: ObserveValueChange, Target: "#input"},
		}},
		{"fill value change", "/actions", Probe{
			Action:  ProbeAction{Kind: ActionFill, Target: "#input", Value: "retained"},
			Observe: ProbeObservation{Kind: ObserveValueChange, Target: "#input"},
		}},
		{"click visible", "/actions", Probe{
			Action:  ProbeAction{Kind: ActionClick, Target: "#show"},
			Observe: ProbeObservation{Kind: ObserveElementVisible, Target: "#hidden"},
		}},
		{"click URL change", "/actions", Probe{
			Action:  ProbeAction{Kind: ActionClick, Target: "#route"},
			Observe: ProbeObservation{Kind: ObserveURLChange},
		}},
		{"pointer text change", "/actions", Probe{
			Action:  ProbeAction{Kind: ActionPointer, Target: "#click"},
			Observe: ProbeObservation{Kind: ObserveTextChange, Target: "#status"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := validateFixture(t, validator, server.URL+tt.path, &tt.probe)
			if report.Status == StatusUnavailable {
				t.Fatalf("configured local Playwright runtime unavailable: %#v", report.Diagnostics)
			}
			if report.Status != StatusPassed || report.Probe == nil || !report.Probe.Passed {
				t.Fatalf("unexpected runtime validation report: %#v", report)
			}
		})
	}
}

func TestPlaywrightRuntimeReportsBehaviorAndPageFailures(t *testing.T) {
	validator := runtimeValidator(t)
	server := serveValidationFixtures(t)
	defer server.Close()

	unchanged := validateFixture(t, validator, server.URL+"/static", &Probe{
		Action:  ProbeAction{Kind: ActionKeyHold, Key: "ArrowRight", DurationMS: 100},
		Observe: ProbeObservation{Kind: ObserveVisualChange, Target: "#game"},
	})
	if unchanged.Status != StatusFailed || unchanged.Probe == nil || unchanged.Probe.Passed {
		t.Fatalf("static output should fail its interaction probe: %#v", unchanged)
	}

	broken := validateFixture(t, validator, server.URL+"/broken", &Probe{
		Action:  ProbeAction{Kind: ActionPointer, Target: "body"},
		Observe: ProbeObservation{Kind: ObserveElementVisible, Target: "body"},
	})
	if broken.Status != StatusFailed {
		t.Fatalf("page errors and failed local assets should fail: %#v", broken)
	}
	failedChecks := map[Check]bool{}
	for _, check := range broken.Checks {
		if !check.Passed {
			failedChecks[check.Check] = true
		}
	}
	if !failedChecks[CheckNoPageErrors] || !failedChecks[CheckNoFailedLocalAsset] {
		t.Fatalf("missing expected failed checks: %#v", broken.Checks)
	}
}

func TestPlaywrightRuntimeUsesBearerAuthentication(t *testing.T) {
	validator := runtimeValidator(t)
	t.Setenv("MYCELIS_API_KEY", "runtime-validator-secret")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer runtime-validator-secret" {
			http.Error(w, "missing authentication token", http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `<!doctype html><button id="run" onclick="document.querySelector('#status').textContent='changed'">Run</button><p id="status">ready</p>`)
	}))
	defer server.Close()

	report := validateFixture(t, validator, server.URL, &Probe{
		Action:  ProbeAction{Kind: ActionClick, Target: "#run"},
		Observe: ProbeObservation{Kind: ObserveTextChange, Target: "#status"},
	})
	if report.Status != StatusPassed || report.Probe == nil || !report.Probe.Passed {
		t.Fatalf("authenticated runtime validation failed: %#v", report)
	}
}
