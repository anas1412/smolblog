package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			if err := buildSite(); err != nil {
				log.Fatalf("Build failed: %v", err)
			}
			fmt.Println("Site built successfully in dist/")
			return
		case "help", "--help", "-h":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	// Ensure the default theme exists
	ensureDefaultTheme()

	// Ensure content directories exist
	if err := os.MkdirAll("content/posts", 0755); err != nil {
		log.Fatalf("Failed to create content directory: %v", err)
	}

	// Auto-build the site so the preview is always up to date
	if err := buildSite(); err != nil {
		log.Printf("Warning: initial build failed: %v", err)
	}

	http.HandleFunc("/admin", handleAdminDashboard)
	http.HandleFunc("/admin/new", handleNewPost)
	http.HandleFunc("/admin/edit", handleEditPost)
	http.HandleFunc("/admin/save", handleSavePost)
	http.HandleFunc("/admin/delete", handleDeletePost)
	http.HandleFunc("/admin/themes", handleThemeManager)
	http.HandleFunc("/admin/themes/upload", handleThemeUpload)
	http.HandleFunc("/admin/themes/activate", handleThemeActivate)
	http.HandleFunc("/admin/themes/export", handleThemeExport)
	http.HandleFunc("/admin/themes/delete", handleThemeDelete)
	http.HandleFunc("/admin/build", handleTriggerBuild)
	http.HandleFunc("/admin/publish", handlePublish)
	http.HandleFunc("/admin/settings", handleSettings)
	http.Handle("/", http.FileServer(http.Dir("dist")))

	log.Println("Admin panel at http://localhost:8080/admin")
	log.Println("Site preview at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func printUsage() {
	fmt.Println(`Smolblog — tiny static site generator

Usage:
  smolblog              Build site + start server (admin + preview)
  smolblog build        Build the static site only (for CI/CD)
  smolblog help         Show this help`)
}
