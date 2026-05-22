package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PostListItem is a summary displayed in the admin post list.
type PostListItem struct {
	Slug  string
	Title string
	Date  string
}

// AdminPageData is passed to the admin template.
type AdminPageData struct {
	Posts       []PostListItem
	Editing     *Post      // non-nil when editing
	Settings    SiteConfig // for settings view
	View        string     // "list", "new", "edit", "settings", "themes"
	Themes      []ThemeMeta
	ActiveTheme string
	ThemeError  string // error message for theme operations
}

// postListItems returns all posts summary for the admin list.
func postListItems() []PostListItem {
	posts := readAllPosts()
	items := make([]PostListItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, PostListItem{
			Slug:  p.Slug,
			Title: p.Title,
			Date:  p.Date.Format("Jan 2, 2006"),
		})
	}
	return items
}

// parseAdminTemplate parses index.html with the shared function map.
func parseAdminTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"joinTags": func(tags []string) string { return strings.Join(tags, ", ") },
	}
	return template.New("index.html").Funcs(funcMap).ParseFiles("index.html", "templates/shared.html")
}

// statusHTML returns a colored status banner for inline status updates.
func statusHTML(msg, color string) string {
	return fmt.Sprintf(`<span class="block mt-2 text-xs font-medium text-%s-700 bg-%s-50 p-2 border border-%s-200 rounded whitespace-pre-wrap">%s</span>`, color, color, color, msg)
}

// friendlyGitError converts git error output into user-friendly instructions.
func friendlyGitError(err error, output string) string {
	outputLower := strings.ToLower(output)

	if strings.Contains(err.Error(), "executable file not found") {
		return `Git is not installed.

Install it:
  sudo apt install git        # Ubuntu/Debian
  brew install git            # macOS
  https://git-scm.com/downloads

Or push your code manually from the terminal.`
	}
	if strings.Contains(outputLower, "not a git repository") {
		return `This folder isn't connected to Git yet.

1. Create a repo on GitHub (don't add a README)
2. Run these in your terminal:

   git init
   git add -A
   git commit -m "First commit"
   git remote add origin https://github.com/YOUR_USER/YOUR_REPO.git
   git push -u origin main

Then come back and click Publish.`
	}
	if strings.Contains(outputLower, "no remote") || strings.Contains(outputLower, "does not appear to be a git repository") {
		return `No GitHub remote is configured.

Run this in your terminal:

  git remote add origin https://github.com/YOUR_USER/YOUR_REPO.git
  git push -u origin main

Then click Publish again.`
	}
	if strings.Contains(outputLower, "non-fast-forward") || strings.Contains(outputLower, "rejected") {
		return `Push was rejected — the remote has changes you don't have locally.

Fix it:
  git pull --rebase
  # resolve any conflicts, then:
  git push

Then click Publish again.`
	}
	return fmt.Sprintf("Something went wrong with Git:\n%s\n\nIf you're stuck, try pushing manually.", strings.TrimSpace(output))
}

// renderAdmin renders the admin template with the given data.
func renderAdmin(w http.ResponseWriter, data AdminPageData) {
	tmpl, err := parseAdminTemplate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

// ---------------------------------------------------------------------------
// GET handlers
// ---------------------------------------------------------------------------

// GET /admin — post list (default view).
func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	renderAdmin(w, AdminPageData{
		Posts: postListItems(),
		View:  "list",
	})
}

// GET /admin/new — new post form.
func handleNewPost(w http.ResponseWriter, r *http.Request) {
	renderAdmin(w, AdminPageData{
		Posts: postListItems(),
		View:  "new",
	})
}

// GET /admin/edit — edit post form (requires ?slug=xxx).
func handleEditPost(w http.ResponseWriter, r *http.Request) {
	data := AdminPageData{
		Posts: postListItems(),
		View:  "edit",
	}
	if slug := r.URL.Query().Get("slug"); slug != "" {
		post, err := readPost(slug)
		if err == nil {
			data.Editing = &post
		}
	}
	renderAdmin(w, data)
}

// GET/POST /admin/settings — display settings form or save it.
func handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleSaveSettings(w, r)
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}
	renderAdmin(w, AdminPageData{
		Posts:    postListItems(),
		View:     "settings",
		Settings: cfg,
	})
}

// ---------------------------------------------------------------------------
// POST handlers (form actions)
// ---------------------------------------------------------------------------

// POST /admin/save — create or update a post, auto-build, redirect to list.
func handleSavePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	dateStr := strings.TrimSpace(r.FormValue("date"))
	body := r.FormValue("content")
	originalSlug := strings.TrimSpace(r.FormValue("original_slug"))

	if title == "" || slug == "" || dateStr == "" {
		http.Error(w, "Title, Slug, and Date are required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	var tags []string
	if tagsStr := strings.TrimSpace(r.FormValue("tags")); tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			if tag := strings.TrimSpace(t); tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	post := Post{
		Title: title,
		Date:  Date{Time: date},
		Slug:  slug,
		Tags:  tags,
		Body:  body,
	}

	raw, err := renderFrontmatter(post, post.Body)
	if err != nil {
		http.Error(w, "Failed to render frontmatter", http.StatusInternalServerError)
		return
	}

	if originalSlug != "" && originalSlug != slug {
		os.Remove(filepath.Join("content/posts", originalSlug+".md"))
	}

	filePath := filepath.Join("content/posts", slug+".md")
	if err := ioutil.WriteFile(filePath, raw, 0644); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	buildSite()
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// POST /admin/delete — delete a post, rebuild, redirect to list.
func handleDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimSpace(r.FormValue("slug"))
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	if err := os.Remove(filepath.Join("content/posts", slug+".md")); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	buildSite()
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// POST /admin/settings — save settings, rebuild, redirect to settings.
func handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	footer := strings.TrimSpace(r.FormValue("footer"))
	aboutContent := r.FormValue("about_content")
	homeContent := r.FormValue("home_content")

	if title == "" {
		title = "My Site"
	}
	if footer == "" {
		footer = "Generated statically via Go."
	}

	cfg := SiteConfig{
		Title:        title,
		Footer:       footer,
		AboutContent: aboutContent,
		HomeContent:  homeContent,
	}

	if err := saveConfig(cfg); err != nil {
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}
	if err := cfg.syncPages(); err != nil {
		http.Error(w, "Settings saved, but pages failed to sync", http.StatusInternalServerError)
		return
	}

	buildSite()
	http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
}

// ---------------------------------------------------------------------------
// POST handlers (htmx inline responses)
// ---------------------------------------------------------------------------

// POST /admin/build — build and show status inline.
func handleTriggerBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := buildSite(); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(statusHTML(fmt.Sprintf("Build failed: %v", err), "red")))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(statusHTML("Build completed ✓", "emerald")))
}

// POST /admin/publish — build, commit, push, show status inline.
func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := buildSite(); err != nil {
		w.Write([]byte(statusHTML(fmt.Sprintf("Build failed: %v", err), "red")))
		return
	}

	cmd := exec.Command("git", "add", "-A")
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.Write([]byte(statusHTML(friendlyGitError(err, string(out)), "red")))
		return
	}

	cmd = exec.Command("git", "commit", "-m", "Update site")
	out, err = cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(out), "nothing to commit") {
		w.Write([]byte(statusHTML(friendlyGitError(err, string(out)), "red")))
		return
	}

	cmd = exec.Command("git", "push")
	out, err = cmd.CombinedOutput()
	if err != nil {
		w.Write([]byte(statusHTML(friendlyGitError(err, string(out)), "red")))
		return
	}

	w.Write([]byte(statusHTML("Published to GitHub ✓", "emerald")))
}

// ---------------------------------------------------------------------------
// Theme management handlers
// ---------------------------------------------------------------------------

// GET /admin/themes — theme manager page.
func handleThemeManager(w http.ResponseWriter, r *http.Request) {
	cfg, _ := loadConfig()
	renderAdmin(w, AdminPageData{
		Posts:       postListItems(),
		View:        "themes",
		Settings:    cfg,
		Themes:      listThemes(),
		ActiveTheme: getActiveTheme(cfg),
	})
}

// POST /admin/themes/upload — install a theme from a zip upload.
func handleThemeUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("theme")
	if err != nil {
		http.Error(w, "Missing theme file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := installThemeFromReader(file); err != nil {
		http.Error(w, fmt.Sprintf("Install failed: %v", err), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/themes", http.StatusSeeOther)
}

// POST /admin/themes/activate — switch the active theme.
func handleThemeActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	themeName := strings.TrimSpace(r.FormValue("theme"))
	if themeName == "" {
		http.Error(w, "Theme name is required", http.StatusBadRequest)
		return
	}
	if !themeExists(themeName) {
		http.Error(w, fmt.Sprintf("Theme %q not found", themeName), http.StatusNotFound)
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	cfg.Theme = themeName
	if err := saveConfig(cfg); err != nil {
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}

	// Rebuild the site with the new theme
	buildSite()

	http.Redirect(w, r, "/admin/themes", http.StatusSeeOther)
}

// GET /admin/themes/export?name=X — download a theme as a zip file.
func handleThemeExport(w http.ResponseWriter, r *http.Request) {
	themeName := r.URL.Query().Get("name")
	if themeName == "" {
		http.Error(w, "Theme name is required", http.StatusBadRequest)
		return
	}
	if !themeExists(themeName) {
		http.Error(w, fmt.Sprintf("Theme %q not found", themeName), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-theme.zip"`, themeName))

	if err := exportTheme(themeName, w); err != nil {
		// Can't write error to headers at this point, so log it
		log.Printf("Error exporting theme %q: %v", themeName, err)
	}
}

// POST /admin/themes/delete — delete a theme.
func handleThemeDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	themeName := strings.TrimSpace(r.FormValue("theme"))
	if themeName == "" {
		http.Error(w, "Theme name is required", http.StatusBadRequest)
		return
	}

	// Don't allow deleting the active theme
	cfg, _ := loadConfig()
	if getActiveTheme(cfg) == themeName {
		http.Error(w, "Cannot delete the active theme. Switch to another theme first.", http.StatusBadRequest)
		return
	}

	if err := deleteTheme(themeName); err != nil {
		http.Error(w, fmt.Sprintf("Delete failed: %v", err), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/themes", http.StatusSeeOther)
}
