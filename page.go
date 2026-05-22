package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// Page represents a static content page with YAML frontmatter.
type Page struct {
	Title string `yaml:"title"`

	Slug     string          `yaml:"-"` // derived from filename (set after parse)
	Body     string          `yaml:"-"` // raw markdown
	HTMLBody template.HTML   `yaml:"-"` // rendered markdown (set during build)
}

// pageSlugToActive maps a page filename slug to the navbar active state.
func pageSlugToActive(slug string) string {
	if slug == "index" {
		return "home"
	}
	return slug
}

// pageSlugToOutput maps a page filename slug to its output path.
func pageSlugToOutput(slug string) string {
	if slug == "index" {
		return "dist/index.html"
	}
	return filepath.Join("dist", slug, "index.html")
}

// readAllPages reads every .md file from content/pages/ and returns them
// with the Slug field set to the filename stem.
func readAllPages() []Page {
	files, err := ioutil.ReadDir("content/pages")
	if err != nil {
		return nil
	}

	var pages []Page
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".md" {
			continue
		}
		raw, err := ioutil.ReadFile(filepath.Join("content/pages", f.Name()))
		if err != nil {
			log.Printf("Warning: skipping unreadable page %s: %v", f.Name(), err)
			continue
		}
		var page Page
		body, err := parseFrontmatter(raw, &page)
		if err != nil {
			log.Printf("Warning: skipping page %s: %v", f.Name(), err)
			continue
		}
		page.Body = body
		page.Slug = strings.TrimSuffix(f.Name(), ".md")
		pages = append(pages, page)
	}

	return pages
}
