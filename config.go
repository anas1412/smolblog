package main

import (
	"encoding/json"
	"os"
)

// SiteConfig holds global site settings stored in settings.json.
type SiteConfig struct {
	Title        string `json:"title"`
	Footer       string `json:"footer"`
	Theme        string `json:"theme"`
	BaseURL      string `json:"base_url"`
	AboutContent string `json:"about_content"`
	HomeContent  string `json:"home_content"`
}

func defaultConfig() SiteConfig {
	return SiteConfig{
		Title:   "Smolblog",
		Footer:  "Built with Smolblog — a tiny Go + HTMX static site generator.",
		Theme:   "default",
		BaseURL: "",
		AboutContent: `## About Me

I'm a developer building things with Go, HTMX, and Tailwind CSS. This site is a static site generator I built from scratch.

### How this site works

- Posts are written in Markdown with YAML frontmatter
- The Go SSG compiles them into flat HTML
- Deployed to GitHub Pages via GitHub Actions
- No databases, no servers, no JavaScript frameworks`,
		HomeContent: `# Welcome

This is my corner of the web. I write about things that interest me — code, design, and ideas.

Head over to the [blog](/blog/) to read the latest posts.`,
	}
}

// loadConfig reads settings.json, creating it with defaults if missing.
func loadConfig() (SiteConfig, error) {
	data, err := os.ReadFile("settings.json")
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			cfg.syncPages()
			saveConfig(cfg)
			return cfg, nil
		}
		return SiteConfig{}, err
	}

	var cfg SiteConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig(), nil
	}
	if cfg.Title == "" {
		cfg.Title = defaultConfig().Title
	}
	return cfg, nil
}

func saveConfig(cfg SiteConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("settings.json", data, 0644)
}

// syncPages writes the config's content back to the markdown page files
// so the regular build pipeline picks them up.
func (cfg SiteConfig) syncPages() error {
	// About page
	aboutMD := "---\ntitle: About\n---\n\n" + cfg.AboutContent
	if err := os.WriteFile("content/pages/about.md", []byte(aboutMD), 0644); err != nil {
		return err
	}

	// Home page
	homeMD := "---\ntitle: Home\n---\n\n" + cfg.HomeContent
	if err := os.WriteFile("content/pages/index.md", []byte(homeMD), 0644); err != nil {
		return err
	}

	return nil
}
