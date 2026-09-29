package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// AUTH-C1b settings: one serialized read-modify-write, atomic file writes, a
// corrupt/unreadable file or no settings path is a 503 blocker (nothing
// written, never a silent default). Ported from the AUTHC1-QA race probes.

const authc1bStrictSeed = `{"theme":"aero-light","assistant_name":"Soma","role":"owner","cost_sensitivity":"high","review_strictness":"strict","automation_tolerance":"cautious","escalation_preference":"halt"}`

func authc1bFileSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("settings file is not JSON: %v (%q)", err, raw)
	}
	return m
}

func authc1bPolicyJSON(who *RequestIdentity) string {
	b, _ := json.Marshal(authc1PolicyFor(who))
	return string(b)
}

func authc1bReseed(t *testing.T, path, seed string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAuthC1bSettingsParallelPersonalPutsKeepPolicy(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)
	path := authc1SettingsFile(t, authc1bStrictSeed)
	var downgraded, failed int
	for round := 0; round < 40; round++ {
		authc1bReseed(t, path, authc1bStrictSeed)
		var wg sync.WaitGroup
		var mu sync.Mutex
		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < 10; i++ {
					status, resp := authc1PutSettings(t, s, fmt.Sprintf(`{"theme":"c1b-%d-%d"}`, g, i), standardUserIdentity())
					var m map[string]any
					mu.Lock()
					if status != http.StatusOK {
						failed++
					} else if json.Unmarshal([]byte(resp), &m) == nil && m["review_strictness"] != "strict" {
						downgraded++
					}
					mu.Unlock()
				}
			}(g)
		}
		wg.Wait()
		if gov := authc1bFileSettings(t, path); gov["review_strictness"] != "strict" || gov["automation_tolerance"] != "cautious" || gov["escalation_preference"] != "halt" {
			t.Fatalf("round %d: parallel non-admin personal PUTs reset org policy to %v", round, gov)
		}
	}
	if downgraded > 0 || failed > 0 {
		t.Fatalf("personal PUTs: %d returned downgraded policy, %d failed", downgraded, failed)
	}
}

func TestAuthC1bSettingsTwoConcurrentPutsKeepPolicy(t *testing.T) {
	dbOpt, _ := withDB(t)
	s := newTestServer(dbOpt)
	path := authc1SettingsFile(t, authc1bStrictSeed)
	strictPolicy := authc1bPolicyJSON(otherUserIdentity())
	resets, flips, rounds := 0, 0, 300
	for round := 0; round < rounds; round++ {
		authc1bReseed(t, path, authc1bStrictSeed)
		var wg sync.WaitGroup
		wg.Add(3)
		for g := 0; g < 2; g++ {
			go func(g int) {
				defer wg.Done()
				authc1PutSettings(t, s, fmt.Sprintf(`{"theme":"par-%d-%d"}`, round, g), standardUserIdentity())
			}(g)
		}
		go func() {
			defer wg.Done()
			if authc1bPolicyJSON(otherUserIdentity()) != strictPolicy {
				flips++
			}
		}()
		wg.Wait()
		if authc1bFileSettings(t, path)["review_strictness"] != "strict" {
			resets++
		}
	}
	t.Logf("2 concurrent non-admin theme PUTs: org policy reset in %d/%d rounds; bystander policy differed in %d/%d rounds", resets, rounds, flips, rounds)
	if resets > 0 || flips > 0 {
		t.Fatalf("non-admin personal PUTs changed the approval policy (resets %d, flips %d)", resets, flips)
	}
}

// Admin and personal writes never lose each other's changes.
func TestAuthC1bSettingsAdminAndPersonalChangesBothKept(t *testing.T) {
	lost, rounds := 0, 200
	for round := 0; round < rounds; round++ {
		dbOpt, mock := withDB(t)
		mock.MatchExpectationsInOrder(false)
		expectAudits(mock, 2)
		s := newTestServer(dbOpt)
		path := authc1SettingsFile(t, authc1SettingsSeed)
		var wg sync.WaitGroup
		var adminStatus, personalStatus int
		wg.Add(2)
		go func() {
			defer wg.Done()
			adminStatus, _ = authc1PutSettings(t, s, `{"review_strictness":"strict"}`, adminWithScopes(scopeGovernanceWrite))
		}()
		go func() {
			defer wg.Done()
			personalStatus, _ = authc1PutSettings(t, s, `{"theme":"c1b-theme"}`, standardUserIdentity())
		}()
		wg.Wait()
		got := authc1bFileSettings(t, path)
		if adminStatus != http.StatusOK || personalStatus != http.StatusOK {
			t.Fatalf("round %d: statuses admin=%d personal=%d", round, adminStatus, personalStatus)
		}
		if got["review_strictness"] != "strict" || got["theme"] != "c1b-theme" {
			lost++
		}
	}
	if lost > 0 {
		t.Fatalf("an applied admin change or a saved personal change was lost in %d/%d rounds", lost, rounds)
	}
}

func authc1bAssertSettingsBlocker(t *testing.T, status int, body string) {
	t.Helper()
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status %d want 503: %s", status, body)
	}
	if data := authc1BlockerData(t, body); data["code"] != codeSettingsStoreUnavailable {
		t.Fatalf("code = %v, want %s", data["code"], codeSettingsStoreUnavailable)
	}
}

func TestAuthC1bSettingsCorruptOrUnreadableFileFailsClosed(t *testing.T) {
	for name, who := range map[string]*RequestIdentity{
		"non-admin personal": standardUserIdentity(),
		"admin governance":   adminWithScopes(scopeGovernanceWrite),
	} {
		for kind, seed := range map[string]string{"truncated": `{"theme":"aero-li`, "empty": ``, "not an object": `[1,2]`} {
			t.Run(name+"/"+kind, func(t *testing.T) {
				dbOpt, mock := withDB(t) // no audit expectations: any audit insert fails
				s := newTestServer(dbOpt)
				path := authc1SettingsFile(t, seed)
				status, body := authc1PutSettings(t, s, `{"theme":"c1b","review_strictness":"strict"}`, who)
				authc1bAssertSettingsBlocker(t, status, body)
				if authc1ReadFile(t, path) != seed {
					t.Fatalf("a refused PUT rewrote the corrupt settings file")
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
				rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleUserSettings), http.MethodGet, "/api/v1/user/settings", "", who)
				authc1bAssertSettingsBlocker(t, rr.Code, rr.Body.String())
			})
		}
	}
	t.Run("unreadable path", func(t *testing.T) {
		dir := t.TempDir() // a directory where the file should be
		t.Setenv("MYCELIS_USER_SETTINGS_PATH", dir)
		status, body := authc1PutSettings(t, newTestServer(), `{"theme":"c1b"}`, standardUserIdentity())
		authc1bAssertSettingsBlocker(t, status, body)
	})
}

func TestAuthC1bSettingsNoPathIsBlocker(t *testing.T) {
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", "")
	t.Setenv("HOME", "")
	if p := userSettingsPath(); p != "" {
		t.Skipf("settings path still resolves to %q", p)
	}
	for name, body := range map[string]string{"admin governance": `{"review_strictness":"strict"}`, "personal": `{"theme":"c1b"}`} {
		t.Run(name, func(t *testing.T) {
			dbOpt, mock := withDB(t)
			s := newTestServer(dbOpt)
			status, resp := authc1PutSettings(t, s, body, adminWithScopes(scopeGovernanceWrite))
			authc1bAssertSettingsBlocker(t, status, resp)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// First boot: a missing file is the defaults, and a PUT creates it atomically.
func TestAuthC1bSettingsMissingFileSavesAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "user-settings.json")
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", path)
	status, body := authc1PutSettings(t, newTestServer(), `{"theme":"c1b-first"}`, standardUserIdentity())
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	got := authc1bFileSettings(t, path)
	if got["theme"] != "c1b-first" || got["review_strictness"] != "standard" {
		t.Fatalf("saved %v", got)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temp files left behind: %v", entries)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("settings file stat = %v, %v; want mode 0644", info, err)
	}
}
