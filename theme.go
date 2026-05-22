package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------------
// Theme model
// ---------------------------------------------------------------------------

// ThemeMeta holds the metadata from a theme's theme.json.
type ThemeMeta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

// themePath returns the directory path for a given theme name.
func themePath(name string) string {
	return filepath.Join("themes", name)
}

// themeMetaPath returns the path to a theme's metadata file.
func themeMetaPath(name string) string {
	return filepath.Join("themes", name, "theme.json")
}

// themeTemplatesPath returns the path to a theme's templates directory.
func themeTemplatesPath(name string) string {
	return filepath.Join("themes", name, "templates")
}

// themeStaticPath returns the path to a theme's static assets directory.
func themeStaticPath(name string) string {
	return filepath.Join("themes", name, "static")
}

// ---------------------------------------------------------------------------
// Listing & loading
// ---------------------------------------------------------------------------

// listThemes scans the themes directory and returns metadata for all installed themes.
func listThemes() []ThemeMeta {
	entries, err := ioutil.ReadDir("themes")
	if err != nil {
		return nil
	}

	var themes []ThemeMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := loadThemeMeta(entry.Name())
		if err != nil {
			log.Printf("Warning: skipping theme %s: %v", entry.Name(), err)
			continue
		}
		themes = append(themes, meta)
	}
	return themes
}

// loadThemeMeta reads the theme.json for a given theme name.
func loadThemeMeta(name string) (ThemeMeta, error) {
	data, err := ioutil.ReadFile(themeMetaPath(name))
	if err != nil {
		return ThemeMeta{}, fmt.Errorf("reading theme.json for %q: %w", name, err)
	}

	var meta ThemeMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return ThemeMeta{}, fmt.Errorf("parsing theme.json for %q: %w", name, err)
	}
	if meta.Name == "" {
		meta.Name = name
	}
	if meta.Version == "" {
		meta.Version = "1.0.0"
	}
	return meta, nil
}

// themeExists checks if a theme directory exists.
func themeExists(name string) bool {
	info, err := os.Stat(themePath(name))
	return err == nil && info.IsDir()
}

// ---------------------------------------------------------------------------
// Install
// ---------------------------------------------------------------------------

// installThemeFromReader reads a zip from a reader and installs it as a theme.
// The zip must contain a theme.json at its root (or in a single top-level dir).
func installThemeFromReader(r io.Reader) error {
	// Save upload to a temp file (zip.NewReader needs io.ReaderAt)
	tmpFile, err := ioutil.TempFile("", "theme-upload-*.zip")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, r); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	// Open as zip
	zr, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("reading zip archive: %w", err)
	}
	defer zr.Close()

	// Discover the theme name by finding theme.json in the archive
	themeName := discoverThemeName(zr)
	if themeName == "" {
		return fmt.Errorf("no theme.json found in zip — each theme must include a theme.json")
	}
	if themeName == "default" {
		return fmt.Errorf("cannot overwrite the default theme")
	}

	// If theme already exists, remove it first for clean replacement
	if themeExists(themeName) {
		if err := os.RemoveAll(themePath(themeName)); err != nil {
			return fmt.Errorf("removing existing theme %q: %w", themeName, err)
		}
	}

	// Determine the common root prefix in the zip (if any)
	// e.g., "my-theme/" in "my-theme/templates/base.html"
	rootPrefix := detectRootPrefix(zr)

	// Extract all files
	for _, f := range zr.File {
		// Strip the root prefix to get the relative path within the theme
		relPath := strings.TrimPrefix(f.Name, rootPrefix)
		relPath = strings.TrimPrefix(relPath, "/")

		if relPath == "" {
			continue
		}

		outPath := filepath.Join(themePath(themeName), relPath)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", outPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening %s in zip: %w", f.Name, err)
		}

		out, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("creating %s: %w", outPath, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
		rc.Close()
		out.Close()
	}

	// Verify the theme is valid
	if _, err := loadThemeMeta(themeName); err != nil {
		// Clean up on failure
		os.RemoveAll(themePath(themeName))
		return fmt.Errorf("installed theme is invalid: %w", err)
	}

	return nil
}

// discoverThemeName finds the theme name by looking for theme.json in the zip.
func discoverThemeName(zr *zip.ReadCloser) string {
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "theme.json") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, _ := ioutil.ReadAll(rc)
			rc.Close()

			var meta ThemeMeta
			if json.Unmarshal(data, &meta) == nil && meta.Name != "" {
				return meta.Name
			}
			// Fallback: use the directory containing theme.json
			dir := filepath.Dir(f.Name)
			if dir != "." && dir != "/" {
				return filepath.Base(dir)
			}
			return ""
		}
	}
	return ""
}

// detectRootPrefix finds the common top-level directory in the zip entries.
// Returns "" if files are at root, or "somedir/" if all share that prefix.
func detectRootPrefix(zr *zip.ReadCloser) string {
	var prefix string
	for _, f := range zr.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) < 2 {
			return "" // file at root, no prefix
		}
		if prefix == "" {
			prefix = parts[0] + "/"
		} else if parts[0]+"/" != prefix {
			return "" // inconsistent prefix
		}
	}
	return prefix
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

// exportTheme creates a zip of the theme directory and writes it to w.
func exportTheme(name string, w io.Writer) error {
	if !themeExists(name) {
		return fmt.Errorf("theme %q not found", name)
	}

	zw := zip.NewWriter(w)
	defer zw.Close()

	baseDir := themePath(name)

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		if info.IsDir() {
			header.Name += "/"
		}

		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})

	return err
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// deleteTheme removes a theme directory.
func deleteTheme(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default theme")
	}
	if !themeExists(name) {
		return fmt.Errorf("theme %q not found", name)
	}
	return os.RemoveAll(themePath(name))
}

// ---------------------------------------------------------------------------
// Default theme initializer (called at startup)
// ---------------------------------------------------------------------------

// ensureDefaultTheme makes sure the themes/default directory exists with all
// required files. Copies from the legacy templates/ directory if needed.
func ensureDefaultTheme() {
	if themeExists("default") {
		// Verify it has theme.json
		if _, err := loadThemeMeta("default"); err == nil {
			return
		}
	}

	// Create the default theme from the legacy templates directory
	dirs := []string{
		themePath("default"),
		themeTemplatesPath("default"),
		themeStaticPath("default"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}

	// Copy existing template files
	entries, err := ioutil.ReadDir("templates")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := ioutil.ReadFile(filepath.Join("templates", entry.Name()))
			if err != nil {
				continue
			}
			ioutil.WriteFile(filepath.Join(themeTemplatesPath("default"), entry.Name()), data, 0644)
		}
	}

	// Write theme.json
	meta := ThemeMeta{
		Name:        "default",
		Version:     "1.0.0",
		Author:      "You",
		Description: "The default built-in theme with Tailwind CSS styling.",
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	ioutil.WriteFile(themeMetaPath("default"), data, 0644)
}

// getActiveTheme returns the name of the active theme, defaulting to "default".
func getActiveTheme(cfg SiteConfig) string {
	if cfg.Theme != "" && themeExists(cfg.Theme) {
		return cfg.Theme
	}
	return "default"
}
