package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// ---------------------------------------------------------------------------
// Template context types
// ---------------------------------------------------------------------------

type pageContext struct {
	Active   string
	Title    string
	HTMLBody template.HTML
}

type postContext struct {
	Active   string
	Title    string
	Date     string
	HTMLBody template.HTML
}

type blogListContext struct {
	Active string
	Title  string
	Posts  []blogPostSummary
}

type blogPostSummary struct {
	Title string
	Date  string
	Slug  string
}

// ---------------------------------------------------------------------------
// Markdown rendering
// ---------------------------------------------------------------------------

func renderMarkdown(body string) template.HTML {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.Linkify,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(body), &buf); err != nil {
		log.Printf("Warning: goldmark conversion failed: %v", err)
		return "<p>Error rendering content.</p>"
	}
	return template.HTML(buf.String())
}

// ---------------------------------------------------------------------------
// Site builder
// ---------------------------------------------------------------------------

// buildSite reads all content and generates the complete static site in dist/.
func buildSite() error {
	// Ensure output directories exist
	for _, d := range []string{"dist/blog", "dist/about", "dist/contact"} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	// Load site config for template rendering
	cfg, _ := loadConfig()

	// Determine active theme
	themeName := getActiveTheme(cfg)
	tmplDir := themeTemplatesPath(themeName)

	// Parse shared base template with config-aware functions
	funcMap := template.FuncMap{
		"siteName":   func() string { return cfg.Title },
		"footerText": func() string { return cfg.Footer },
	}
	baseTmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(
		filepath.Join(tmplDir, "base.html"),
		filepath.Join(tmplDir, "shared.html"),
	)
	if err != nil {
		return fmt.Errorf("parse base template from theme %q: %w", themeName, err)
	}

	// -------------------------------------------------------------------
	// 1. Static pages (Home, About, Contact)
	// -------------------------------------------------------------------
	pageTmpl, err := template.Must(baseTmpl.Clone()).ParseFiles(filepath.Join(tmplDir, "page.html"))
	if err != nil {
		return fmt.Errorf("parse page template: %w", err)
	}

	pages := readAllPages()
	for _, p := range pages {
		htmlBody := renderMarkdown(p.Body)
		outPath := pageSlugToOutput(p.Slug)

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(outPath), err)
		}

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", outPath, err)
		}

		ctx := pageContext{
			Active:   pageSlugToActive(p.Slug),
			Title:    p.Title,
			HTMLBody: htmlBody,
		}

		if err := pageTmpl.ExecuteTemplate(f, "base", ctx); err != nil {
			f.Close()
			return fmt.Errorf("execute %s: %w", outPath, err)
		}
		f.Close()
		log.Printf("Built page: %s", outPath)
	}

	// -------------------------------------------------------------------
	// 2. Blog posts (individual)
	// -------------------------------------------------------------------
	postTmpl, err := template.Must(baseTmpl.Clone()).ParseFiles(filepath.Join(tmplDir, "post.html"))
	if err != nil {
		return fmt.Errorf("parse post template: %w", err)
	}

	posts := readAllPosts()

	var summaries []blogPostSummary

	for _, p := range posts {
		htmlBody := renderMarkdown(p.Body)
		outPath := filepath.Join("dist/blog", p.Slug+".html")

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", outPath, err)
		}

		ctx := postContext{
			Active:   "blog",
			Title:    p.Title,
			Date:     p.Date.Format("January 2, 2006"),
			HTMLBody: htmlBody,
		}

		if err := postTmpl.ExecuteTemplate(f, "base", ctx); err != nil {
			f.Close()
			return fmt.Errorf("execute %s: %w", outPath, err)
		}
		f.Close()

		summaries = append(summaries, blogPostSummary{
			Title: p.Title,
			Date:  p.Date.Format("January 2, 2006"),
			Slug:  p.Slug,
		})
		log.Printf("Built post: %s", outPath)
	}

	// -------------------------------------------------------------------
	// 3. Blog listing page (dist/blog/index.html)
	// -------------------------------------------------------------------
	listTmpl, err := template.Must(baseTmpl.Clone()).ParseFiles(filepath.Join(tmplDir, "blog-list.html"))
	if err != nil {
		return fmt.Errorf("parse blog-list template: %w", err)
	}

	listOut := "dist/blog/index.html"
	f, err := os.Create(listOut)
	if err != nil {
		return fmt.Errorf("create %s: %w", listOut, err)
	}
	defer f.Close()

	if err := listTmpl.ExecuteTemplate(f, "base", blogListContext{
		Active: "blog",
		Title:  "Blog",
		Posts:  summaries,
	}); err != nil {
		return fmt.Errorf("execute %s: %w", listOut, err)
	}
	log.Printf("Built blog listing: %s", listOut)

	// -------------------------------------------------------------------
	// 4. Copy theme static assets to dist/
	// -------------------------------------------------------------------
	if err := copyThemeStatic(themeName); err != nil {
		log.Printf("Warning: copying theme static assets: %v", err)
	}

	return nil
}

// copyThemeStatic copies a theme's static/ directory into dist/.
func copyThemeStatic(themeName string) error {
	staticDir := themeStaticPath(themeName)
	info, err := os.Stat(staticDir)
	if err != nil || !info.IsDir() {
		return nil // no static dir, nothing to copy
	}

	return filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		dst := filepath.Join("dist", relPath)

		if info.IsDir() {
			return os.MkdirAll(dst, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}

		return os.WriteFile(dst, data, 0644)
	})
}
