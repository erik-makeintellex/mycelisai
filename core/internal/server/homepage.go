package server

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
	"gopkg.in/yaml.v3"
)

const defaultHomepageConfigPath = "/core/config/homepage.yaml"

type HomepageConfig struct {
	Enabled      bool              `json:"enabled" yaml:"enabled"`
	Brand        HomepageBrand     `json:"brand" yaml:"brand"`
	Hero         HomepageHero      `json:"hero" yaml:"hero"`
	Announcement HomepageBanner    `json:"announcement,omitempty" yaml:"announcement"`
	Sections     []HomepageSection `json:"sections" yaml:"sections"`
	Links        []HomepageLink    `json:"links" yaml:"links"`
	FooterText   string            `json:"footer_text,omitempty" yaml:"footer_text"`
	ConfigIssue  string            `json:"config_issue,omitempty"`
}

type HomepageBrand struct {
	ProductName string `json:"product_name" yaml:"product_name"`
	Tagline     string `json:"tagline" yaml:"tagline"`
	LogoURL     string `json:"logo_url,omitempty" yaml:"logo_url"`
}

type HomepageHero struct {
	Headline     string      `json:"headline" yaml:"headline"`
	Subheadline  string      `json:"subheadline" yaml:"subheadline"`
	PrimaryCTA   HomepageCTA `json:"primary_cta" yaml:"primary_cta"`
	SecondaryCTA HomepageCTA `json:"secondary_cta" yaml:"secondary_cta"`
}

type HomepageCTA struct {
	Label    string `json:"label" yaml:"label"`
	Href     string `json:"href" yaml:"href"`
	External bool   `json:"external,omitempty" yaml:"external"`
}

type HomepageBanner struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Text    string `json:"text" yaml:"text"`
}

type HomepageSection struct {
	Title string `json:"title" yaml:"title"`
	Body  string `json:"body" yaml:"body"`
}

type HomepageLink struct {
	Label       string `json:"label" yaml:"label"`
	Href        string `json:"href" yaml:"href"`
	Description string `json:"description" yaml:"description"`
	External    bool   `json:"external,omitempty" yaml:"external"`
}

type homepageFile struct {
	Homepage HomepageConfig `yaml:"homepage"`
}

func defaultHomepageConfig() HomepageConfig {
	return HomepageConfig{
		Enabled: true,
		Brand: HomepageBrand{
			ProductName: "Mycelis",
			Tagline:     "Soma-centered AI organization orchestration",
		},
		Hero: HomepageHero{
			Headline:    "Operate AI Organizations through Soma",
			Subheadline: "Express intent once. Soma coordinates teams, tools, reviews, and governed execution behind the scenes.",
			PrimaryCTA:  HomepageCTA{Label: "Start with Soma", Href: "/dashboard"},
			SecondaryCTA: HomepageCTA{
				Label: "View Documentation",
				Href:  "/docs",
			},
		},
		Sections: []HomepageSection{
			{Title: "Express intent", Body: "Describe what you want to accomplish in plain language."},
			{Title: "Soma coordinates", Body: "Soma turns intent into structured work across teams, tools, memory, and reviews."},
			{Title: "Governed execution", Body: "Approvals, audit records, and capability boundaries keep actions controlled."},
		},
		Links: []HomepageLink{
			{Label: "Documentation", Href: "/docs", Description: "Read setup and architecture guidance."},
			{Label: "Resources and MCP", Href: "/resources", Description: "Review connected tools and runtime capabilities."},
			{Label: "Activity and Audit", Href: "/activity", Description: "Inspect system activity and governed execution."},
			{Label: "Status", Href: "/system", Description: "Review service health and recovery status."},
		},
		FooterText: "Self-hosted AI Organization orchestration with Soma.",
	}
}

func homepageConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("MYCELIS_HOMEPAGE_CONFIG_PATH")); path != "" {
		return path
	}
	return defaultHomepageConfigPath
}

func loadHomepageConfig(path string) HomepageConfig {
	cfg := defaultHomepageConfig()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("[homepage] config read failed: %v", err)
			cfg.ConfigIssue = "homepage config unreadable; defaults applied"
		}
		return cfg
	}
	var file homepageFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		log.Printf("[homepage] config parse failed: %v", err)
		cfg.ConfigIssue = "homepage config malformed; defaults applied"
		return cfg
	}
	if !file.Homepage.Enabled {
		return cfg
	}
	if sanitized, ok := sanitizeHomepageConfig(file.Homepage); ok {
		return sanitized
	}
	cfg.ConfigIssue = "homepage config incomplete; defaults applied"
	return cfg
}

func sanitizeHomepageConfig(cfg HomepageConfig) (HomepageConfig, bool) {
	cfg.Brand.ProductName = safeText(cfg.Brand.ProductName)
	cfg.Brand.Tagline = safeText(cfg.Brand.Tagline)
	logoExternal := false
	cfg.Brand.LogoURL = safeLink(cfg.Brand.LogoURL, &logoExternal)
	cfg.Hero.Headline = safeText(cfg.Hero.Headline)
	cfg.Hero.Subheadline = safeText(cfg.Hero.Subheadline)
	cfg.Hero.PrimaryCTA = sanitizeCTA(cfg.Hero.PrimaryCTA)
	cfg.Hero.SecondaryCTA = sanitizeCTA(cfg.Hero.SecondaryCTA)
	cfg.Announcement.Text = safeText(cfg.Announcement.Text)
	cfg.FooterText = safeText(cfg.FooterText)
	cfg.Sections = sanitizeSections(cfg.Sections)
	cfg.Links = sanitizeLinks(cfg.Links)
	if cfg.Brand.ProductName == "" || cfg.Hero.Headline == "" || cfg.Hero.PrimaryCTA.Href == "" {
		return cfg, false
	}
	if len(cfg.Sections) == 0 || len(cfg.Links) == 0 {
		return cfg, false
	}
	return cfg, true
}

func sanitizeCTA(cta HomepageCTA) HomepageCTA {
	cta.Label = safeText(cta.Label)
	cta.Href = safeLink(cta.Href, &cta.External)
	return cta
}

func sanitizeSections(raw []HomepageSection) []HomepageSection {
	out := make([]HomepageSection, 0, len(raw))
	for _, section := range raw {
		title, body := safeText(section.Title), safeText(section.Body)
		if title != "" && body != "" {
			out = append(out, HomepageSection{Title: title, Body: body})
		}
	}
	return out
}

func sanitizeLinks(raw []HomepageLink) []HomepageLink {
	out := make([]HomepageLink, 0, len(raw))
	for _, link := range raw {
		link.Label = safeText(link.Label)
		link.Description = safeText(link.Description)
		link.Href = safeLink(link.Href, &link.External)
		if link.Label != "" && link.Href != "" {
			out = append(out, link)
		}
	}
	return out
}

func safeText(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\x00", "")
	value = strings.ReplaceAll(value, "<", "")
	return strings.ReplaceAll(value, ">", "")
}

func safeLink(href string, external *bool) string {
	href = strings.TrimSpace(href)
	lower := strings.ToLower(href)
	if strings.HasPrefix(href, "/") {
		return href
	}
	if strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "mailto:") {
		*external = true
		return href
	}
	if href == "" {
		return ""
	}
	return "#"
}

func (s *AdminServer) HandleHomepageConfig(w http.ResponseWriter, r *http.Request) {
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(loadHomepageConfig(homepageConfigPath())))
}
