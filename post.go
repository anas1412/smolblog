package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
)

// Post represents a blog post with YAML frontmatter metadata.
type Post struct {
	Title string   `yaml:"title"`
	Date  Date     `yaml:"date"`
	Slug  string   `yaml:"slug"`
	Tags  []string `yaml:"tags,omitempty"`

	Body     string          `yaml:"-"` // raw markdown
	HTMLBody template.HTML   `yaml:"-"` // rendered markdown (set during build)
}

// readPost reads a single post by slug.
func readPost(slug string) (Post, error) {
	raw, err := ioutil.ReadFile(filepath.Join("content/posts", slug+".md"))
	if err != nil {
		return Post{}, err
	}
	var post Post
	body, err := parseFrontmatter(raw, &post)
	if err != nil {
		return Post{}, err
	}
	post.Body = body
	return post, nil
}

// readAllPosts reads every .md file from content/posts/ and returns them
// sorted by date descending (newest first).
func readAllPosts() []Post {
	files, err := ioutil.ReadDir("content/posts")
	if err != nil {
		return nil
	}

	var posts []Post
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".md" {
			continue
		}
		raw, err := ioutil.ReadFile(filepath.Join("content/posts", f.Name()))
		if err != nil {
			log.Printf("Warning: skipping unreadable post %s: %v", f.Name(), err)
			continue
		}
		var post Post
		body, err := parseFrontmatter(raw, &post)
		if err != nil {
			log.Printf("Warning: skipping post %s: %v", f.Name(), err)
			continue
		}
		post.Body = body
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date.Time)
	})

	return posts
}
