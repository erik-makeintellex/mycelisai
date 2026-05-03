package server

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestLoadHomepageConfigDefaultWhenMissing(t *testing.T) {
	cfg := loadHomepageConfig(filepath.Join(t.TempDir(), "missing.yaml"))

	if cfg.Brand.ProductName != "Mycelis" {
		t.Fatalf("expected default product name, got %q", cfg.Brand.ProductName)
	}
	if cfg.Hero.PrimaryCTA.Href != "/dashboard" {
		t.Fatalf("expected default Soma CTA, got %q", cfg.Hero.PrimaryCTA.Href)
	}
}

func TestLoadHomepageConfigCustom(t *testing.T) {
	path := writeHomepageConfig(t, `
homepage:
  enabled: true
  brand:
    product_name: "Acme AI"
    tagline: "Internal governed execution"
  hero:
    headline: "Coordinate work through Soma"
    subheadline: "Soma routes intent into operational teams."
    primary_cta:
      label: "Open Soma"
      href: "/dashboard"
    secondary_cta:
      label: "Support"
      href: "https://support.example.com"
  sections:
    - title: "Express intent"
      body: "Tell Soma what needs to happen."
  links:
    - label: "Runbook"
      href: "https://docs.example.com"
      description: "Internal deployment guidance."
`)

	cfg := loadHomepageConfig(path)
	if cfg.Brand.ProductName != "Acme AI" {
		t.Fatalf("expected custom brand, got %q", cfg.Brand.ProductName)
	}
	if !cfg.Hero.SecondaryCTA.External || !cfg.Links[0].External {
		t.Fatal("expected external links to be marked")
	}
}

func TestLoadHomepageConfigMalformedFallsBack(t *testing.T) {
	path := writeHomepageConfig(t, "homepage:\n  hero: [")

	cfg := loadHomepageConfig(path)
	if cfg.Brand.ProductName != "Mycelis" || cfg.ConfigIssue == "" {
		t.Fatalf("expected safe fallback with issue, got %#v", cfg)
	}
}

func TestLoadHomepageConfigSanitizesUnsafeLinkAndHTML(t *testing.T) {
	path := writeHomepageConfig(t, `
homepage:
  enabled: true
  brand:
    product_name: "<Acme>"
  hero:
    headline: "<Operate>"
    subheadline: "Safe"
    primary_cta:
      label: "Start"
      href: "javascript:alert(1)"
    secondary_cta:
      label: "Docs"
      href: "/docs"
  sections:
    - title: "Intent"
      body: "No html"
  links:
    - label: "Docs"
      href: "/docs"
      description: "Safe"
`)

	cfg := loadHomepageConfig(path)
	if cfg.Brand.ProductName != "Acme" || cfg.Hero.Headline != "Operate" {
		t.Fatalf("expected text sanitization, got %#v", cfg)
	}
	if cfg.Hero.PrimaryCTA.Href != "#" {
		t.Fatalf("expected unsafe link rewritten, got %q", cfg.Hero.PrimaryCTA.Href)
	}
}

func TestHandleHomepageConfig(t *testing.T) {
	path := writeHomepageConfig(t, `
homepage:
  enabled: true
  brand:
    product_name: "Portal"
  hero:
    headline: "Operate through Soma"
    subheadline: "Governed execution"
    primary_cta:
      label: "Start"
      href: "/dashboard"
    secondary_cta:
      label: "Docs"
      href: "/docs"
  sections:
    - title: "Intent"
      body: "Plain language"
  links:
    - label: "Docs"
      href: "/docs"
      description: "Read guidance"
`)
	t.Setenv("MYCELIS_HOMEPAGE_CONFIG_PATH", path)
	server := newTestServer()
	mux := setupMux(t, "GET /api/v1/homepage", server.HandleHomepageConfig)

	rr := doRequest(t, mux, http.MethodGet, "/api/v1/homepage", "")
	assertStatus(t, rr, http.StatusOK)
	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	if !resp.OK {
		t.Fatalf("expected ok response: %#v", resp)
	}
}

func writeHomepageConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "homepage.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
