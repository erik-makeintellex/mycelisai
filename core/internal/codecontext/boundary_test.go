package codecontext

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const leakMarker = "LEAKMARKER"

// boundaryFixture builds a source root holding normal source files and
// sensitive files that all contain leakMarker, plus an outside directory.
func boundaryFixture(t *testing.T) (*Service, string, string) {
	t.Helper()
	root := t.TempDir()
	outside := t.TempDir()
	writeFile(t, filepath.Join(root, "core", "widget.go"), "package core\n\nfunc BuildWidget() string { return \"ok\" }\n")
	writeFile(t, filepath.Join(root, "ui", "widget.ts"), "export function buildWidget() { return 'ok' }\n")
	writeFile(t, filepath.Join(root, "docs", "widget.md"), "# BuildWidget guide\n")
	writeFile(t, filepath.Join(root, ".env"), "API_KEY="+leakMarker+"\n")
	writeFile(t, filepath.Join(root, ".env.local"), "API_KEY="+leakMarker+"\n")
	writeFile(t, filepath.Join(root, ".git"), "gitdir: "+outside+"/"+leakMarker+"/worktrees/x\n")
	writeFile(t, filepath.Join(root, "credentials.json"), "{\"key\":\""+leakMarker+"\"}\n")
	writeFile(t, filepath.Join(root, "deploy", "kubeconfig.yaml"), "token: "+leakMarker+"\n")
	writeFile(t, filepath.Join(root, "deploy", "service-account.json"), "{\"k\":\""+leakMarker+"\"}\n")
	writeFile(t, filepath.Join(root, "certs", "server.pem"), leakMarker+"\n")
	writeFile(t, filepath.Join(root, "keys", "id_ed25519"), leakMarker+"\n")
	writeFile(t, filepath.Join(root, ".npmrc"), "//registry/:_authToken="+leakMarker+"\n")
	writeFile(t, filepath.Join(outside, "outside.go"), "package outside\n\nconst Value = \""+leakMarker+"\"\n")
	return NewService(Config{SourceRoots: []string{root}}), root, outside
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(raw)
}

// assertNoLeak fails when a response carries the marker or an absolute host
// path. Query and Target only echo the caller's own input, so they are cleared.
func assertNoLeak(t *testing.T, resp Response, hostPaths ...string) {
	t.Helper()
	resp.Query, resp.Target = "", ""
	body := mustJSON(t, resp)
	if strings.Contains(body, leakMarker) {
		t.Fatalf("response leaked sensitive content: %s", body)
	}
	for _, p := range hostPaths {
		if strings.Contains(body, p) || strings.Contains(body, filepath.ToSlash(p)) {
			t.Fatalf("response leaked absolute host path %q: %s", p, body)
		}
	}
}

func assertPathUnavailable(t *testing.T, resp Response) {
	t.Helper()
	if resp.Status != "blocked" || resp.Blocker == nil || resp.Count != 0 || len(resp.Refs) != 0 || len(resp.ExtractedFacts) != 0 {
		t.Fatalf("expected blocked response without refs, got %s", mustJSON(t, resp))
	}
	if resp.Blocker.Message != errPathUnavailable.Error() {
		t.Fatalf("blocker message must be the generic not-available message, got %q", resp.Blocker.Message)
	}
}

func TestQueryExcludedFileSubpathsAreBlocked(t *testing.T) {
	svc, root, outside := boundaryFixture(t)
	missing, err := svc.Query(context.Background(), Request{Query: "API", Path: "does-not-exist.go"})
	if err != nil {
		t.Fatalf("Query missing: %v", err)
	}
	assertPathUnavailable(t, missing)
	for _, p := range []string{".env", "./.env", ".env.local", ".git", ".git/config", "credentials.json", "deploy/kubeconfig.yaml", "deploy/service-account.json", "certs/server.pem", "keys/id_ed25519", ".npmrc", "deploy"} {
		t.Run(p, func(t *testing.T) {
			resp, err := svc.Query(context.Background(), Request{Query: leakMarker, Path: p})
			if err != nil {
				t.Fatalf("Query: %v", err)
			}
			if p == "deploy" {
				if resp.Status != "ok" || resp.Count != 0 {
					t.Fatalf("directory holding only excluded files must return no refs: %s", mustJSON(t, resp))
				}
			} else {
				assertPathUnavailable(t, resp)
				if *resp.Blocker != *missing.Blocker {
					t.Fatalf("excluded path must look like a missing path: %+v vs %+v", resp.Blocker, missing.Blocker)
				}
			}
			assertNoLeak(t, resp, root, outside)
		})
	}
}

func TestQueryWholeRootNeverIndexesSensitiveFiles(t *testing.T) {
	svc, root, outside := boundaryFixture(t)
	for _, q := range []string{leakMarker, "gitdir", "kubeconfig", "credentials", ".env", "id_ed25519"} {
		resp, err := svc.Query(context.Background(), Request{Query: q, Limit: 25})
		if err != nil {
			t.Fatalf("Query %q: %v", q, err)
		}
		if resp.Count != 0 {
			t.Fatalf("query %q returned sensitive refs: %s", q, mustJSON(t, resp))
		}
		assertNoLeak(t, resp, root, outside)
	}
	impact, err := svc.Impact(context.Background(), Request{Path: ".env", Target: leakMarker})
	if err != nil {
		t.Fatalf("Impact: %v", err)
	}
	if impact.Count != 0 {
		t.Fatalf("impact returned sensitive refs: %s", mustJSON(t, impact))
	}
	assertNoLeak(t, impact, root, outside)
}

func TestExplainExcludedFilesAreBlocked(t *testing.T) {
	svc, root, outside := boundaryFixture(t)
	for _, p := range []string{".env", ".git", "credentials.json", "deploy/kubeconfig.yaml", "certs/server.pem", "does-not-exist.go", "core", "../outside.go", filepath.Join(outside, "outside.go")} {
		resp, err := svc.Explain(context.Background(), Request{Path: p})
		if err != nil {
			t.Fatalf("Explain %q: %v", p, err)
		}
		assertPathUnavailable(t, resp)
		assertNoLeak(t, resp, root, outside)
	}
}

func TestSymlinkedFileEscapingRootIsRejected(t *testing.T) {
	svc, root, outside := boundaryFixture(t)
	symlinkOrSkip(t, filepath.Join(outside, "outside.go"), filepath.Join(root, "core", "link.go"))
	symlinkOrSkip(t, filepath.Join(root, ".env"), filepath.Join(root, "docs", "alias.md"))

	for _, p := range []string{"core/link.go", "docs/alias.md"} {
		query, err := svc.Query(context.Background(), Request{Query: leakMarker, Path: p})
		if err != nil {
			t.Fatalf("Query %q: %v", p, err)
		}
		assertPathUnavailable(t, query)
		assertNoLeak(t, query, root, outside)
		explain, err := svc.Explain(context.Background(), Request{Path: p})
		if err != nil {
			t.Fatalf("Explain %q: %v", p, err)
		}
		assertPathUnavailable(t, explain)
		assertNoLeak(t, explain, root, outside)
	}
	walk, err := svc.Query(context.Background(), Request{Query: leakMarker, Limit: 25})
	if err != nil {
		t.Fatalf("Query walk: %v", err)
	}
	if walk.Count != 0 {
		t.Fatalf("walk followed an escaping symlink: %s", mustJSON(t, walk))
	}
	assertNoLeak(t, walk, root, outside)
}

func TestSymlinkedDirectoryEscapingRootIsRejected(t *testing.T) {
	svc, root, outside := boundaryFixture(t)
	symlinkOrSkip(t, outside, filepath.Join(root, "linked"))

	for _, p := range []string{"linked", "linked/outside.go"} {
		resp, err := svc.Query(context.Background(), Request{Query: leakMarker, Path: p})
		if err != nil {
			t.Fatalf("Query %q: %v", p, err)
		}
		assertPathUnavailable(t, resp)
		assertNoLeak(t, resp, root, outside)
	}
	explain, err := svc.Explain(context.Background(), Request{Path: "linked/outside.go"})
	if err != nil {
		t.Fatalf("Explain: %v", err)
	}
	assertPathUnavailable(t, explain)
	walk, err := svc.Query(context.Background(), Request{Query: leakMarker, Limit: 25})
	if err != nil {
		t.Fatalf("Query walk: %v", err)
	}
	if walk.Count != 0 {
		t.Fatalf("walk followed an escaping directory symlink: %s", mustJSON(t, walk))
	}
	if _, err := svc.RegisterSource(context.Background(), SourceInput{ID: "linked-source", Name: "Linked", SourceType: "repository", RootPath: filepath.Join(root, "linked")}); err == nil {
		t.Fatal("RegisterSource through an escaping symlink succeeded")
	}
}

func TestRegisterSourceRejectsExcludedDirectories(t *testing.T) {
	svc, root, _ := boundaryFixture(t)
	writeFile(t, filepath.Join(root, "secrets", "notes.md"), leakMarker+"\n")
	for _, p := range []string{"secrets", "deploy/../.git"} {
		if _, err := svc.RegisterSource(context.Background(), SourceInput{ID: "bad-source", Name: "Bad", SourceType: "repository", RootPath: filepath.Join(root, p)}); err == nil {
			t.Fatalf("RegisterSource %q succeeded", p)
		}
	}
	empty := NewService(Config{})
	if _, err := empty.RegisterSource(context.Background(), SourceInput{ID: "any-source", Name: "Any", SourceType: "repository", RootPath: root}); err == nil {
		t.Fatal("RegisterSource without configured roots must fail closed")
	}
}

func TestNormalSourceFilesStayIndexedAndQueryable(t *testing.T) {
	svc, root, outside := boundaryFixture(t)
	resp, err := svc.Query(context.Background(), Request{Query: "widget", Limit: 25})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	seen := map[string]bool{}
	for _, ref := range resp.Refs {
		if filepath.IsAbs(ref.FilePath) || strings.HasPrefix(ref.FilePath, "..") {
			t.Fatalf("ref path is not source-relative: %q", ref.FilePath)
		}
		seen[ref.FilePath] = true
	}
	for _, want := range []string{"core/widget.go", "ui/widget.ts", "docs/widget.md"} {
		if !seen[want] {
			t.Fatalf("expected %s to be indexed, refs = %s", want, mustJSON(t, resp))
		}
	}
	assertNoLeak(t, resp, root, outside)
	for _, p := range []string{"core/widget.go", "ui/widget.ts", "docs/widget.md", "core"} {
		sub, err := svc.Query(context.Background(), Request{Query: "widget", Path: p})
		if err != nil || sub.Status != "ok" || sub.Count == 0 {
			t.Fatalf("Query path %q = %s, err %v", p, mustJSON(t, sub), err)
		}
	}
	explain, err := svc.Explain(context.Background(), Request{Path: "core/widget.go"})
	if err != nil || explain.Status != "ok" || len(explain.ExtractedFacts) == 0 {
		t.Fatalf("Explain = %s, err %v", mustJSON(t, explain), err)
	}
	assertNoLeak(t, explain, root, outside)
	index, err := svc.Index(context.Background(), "")
	if err != nil || index.Status != "ok" {
		t.Fatalf("Index = %s, err %v", mustJSON(t, index), err)
	}
	if scanned, _ := index.Metadata["scanned_files"].(int); scanned != 3 {
		t.Fatalf("expected exactly the 3 normal files to be scanned, got %v", index.Metadata["scanned_files"])
	}
}

func TestSensitiveNameRule(t *testing.T) {
	for _, name := range []string{".git", ".env", ".env.production", "server.pem", "tls.key", "client.p12", "client.pfx", "id_rsa", "id_rsa.pub", "id_ecdsa", "id_ed25519", "id_custom", "aws_credentials", "Credentials.json", "kubeconfig", "my-kubeconfig.yaml", "token.json", "client_secret.yaml", "service-account.json", "serviceaccount.yaml", "service_account.json", ".npmrc", ".pypirc", ".netrc"} {
		if !sensitiveName(name) || includeFile(name) {
			t.Fatalf("%q must be excluded", name)
		}
	}
	for _, name := range []string{"main.go", "widget.ts", "README.md", "config.yaml", "package.json", "id_generator.go"} {
		if sensitiveName(name) || !includeFile(name) {
			t.Fatalf("%q must be included", name)
		}
	}
}

func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}
