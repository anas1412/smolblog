# Smolblog

<p>
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white" alt="Go version">
  <img src="https://img.shields.io/badge/HTMX-2.0-3366CC?logo=htmx&logoColor=white" alt="HTMX">
  <img src="https://img.shields.io/badge/Tailwind-4-06B6D4?logo=tailwindcss&logoColor=white" alt="Tailwind CSS">
  <img src="https://img.shields.io/github/license/anas1412/smolblog" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/anas1412/smolblog/deploy.yml?logo=github&label=deploy" alt="GitHub Actions">
</p>

A lightweight static site generator built with Go. Write content in Markdown with YAML frontmatter, generate flat HTML, and deploy to GitHub Pages — no database or server needed.

Built with [Goldmark](https://github.com/yuin/goldmark) for Markdown rendering, [HTMX](https://htmx.org/) for the admin UI, and [Tailwind CSS](https://tailwindcss.com/) for styling.

---

## Features

- **YAML frontmatter** — standard `title`, `date`, `slug`, and `tags` metadata per post
- **Static pages** — Home, Blog, About, Contact with shared navbar
- **Blog listing** — posts sorted by date, newest first
- **Admin panel** — web UI at `/admin` for composing, editing, and managing posts
- **One-click Build** — rebuilds the static site from the admin panel
- **One-click Publish** — commits and pushes to GitHub from the admin panel
- **Live preview** — view your site at any time from the admin
- **Theme system** — install, switch, and export themes from the admin UI (no code needed)
- **CLI build** — `./smolblog build` generates the complete site into `dist/`
- **GitHub Actions** — auto-builds and deploys to Pages on push

---

## Quick Start

### 1. Get the code

**Fork** the repo (recommended) — click the "Fork" button on GitHub, then:

```bash
git clone https://github.com/YOUR_USERNAME/smolblog.git
cd smolblog
```

Or clone and set up fresh:

```bash
git clone https://github.com/anas1412/smolblog.git
cd smolblog
git remote set-url origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
```

### 2. Run locally

Requires [Go 1.26+](https://go.dev/dl/).

```bash
# Build the binary
go build -o smolblog .

# Build the site + start the admin panel and preview server
./smolblog
```

Open [http://localhost:8080/admin](http://localhost:8080/admin) to access the admin panel.
Your site preview is at [http://localhost:8080/](http://localhost:8080/).

### 3. Write a Post

In the admin panel, fill in the form (title, slug, date, tags, markdown body) and click **Save Post**.

Or create a Markdown file manually in `content/posts/`:

```markdown
---
title: My First Post
date: 2026-05-22
slug: my-first-post
tags:
  - go
  - tutorial
---

# Hello

This is the body of my post written in **Markdown**.
```

### 4. Build the Site

```bash
./smolblog build
```

This generates the complete static site into `dist/`:

```
dist/
├── index.html            # Home page
├── blog/
│   ├── index.html        # Blog listing
│   ├── hello-world.html  # Post
│   └── ...
├── about/
│   └── index.html        # About page
└── contact/
    └── index.html        # Contact page
```

---

## Deploy to GitHub Pages

### 1. Push to GitHub

```bash
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
git push -u origin main
```

### 2. Enable GitHub Pages

1. Go to your repo on GitHub → **Settings** → **Pages**
2. Under **Source**, select **GitHub Actions**
3. That's it — the included `.github/workflows/deploy.yml` handles everything

### 3. Every Push Deploys

Every push to `main` triggers the workflow, which:

1. Builds the Go binary
2. Runs `./smolblog build` to generate the site
3. Uploads `dist/` as a Pages artifact
4. Deploys it

Your site will be live at `https://YOUR_USERNAME.github.io/YOUR_REPO/`.

The base URL is **auto-detected** in the workflow — no manual config needed. If your repo is `yourname/your-blog`, it deploys to `/your-blog/`. If it's a user site (`yourname/yourname.github.io`), it deploys to the root.

You can also trigger a manual deploy from the **Actions** tab → **Build and Deploy to GitHub Pages** → **Run workflow**.

### Publish Button (Admin)

The admin's **Publish** button runs `git commit && git push` from the server. For this to work, you need:

- **SSH key** added to your GitHub account — [github.com/settings/keys](https://github.com/settings/keys)
- **Git remote** set to SSH: `git remote set-url origin git@github.com:YOUR_USER/YOUR_REPO.git`

If you haven't set this up, the button will show a helpful error message with next steps.

---

## Content Format

### Posts (`content/posts/`)

```markdown
---
title: Hello World
date: 2026-05-22
slug: hello-world
tags:
  - go
  - meta
---

Markdown body goes here…
```

| Field | Required | Description |
|---|---|---|
| `title` | Yes | Post title (used in `<title>`, blog listing, and post page) |
| `date` | Yes | Publish date in `YYYY-MM-DD` format (posts sort newest-first) |
| `slug` | Yes | URL slug — `/blog/{slug}.html` |
| `tags` | No | Comma or YAML list of tags for the post |

### Pages (`content/pages/`)

```markdown
---
title: About
---

Markdown body goes here…
```

| Field | Required | Description |
|---|---|---|
| `title` | Yes | Page title |

The filename determines the URL:
- `index.md` → `/` (Home, with "Home" active in navbar)
- `about.md` → `/about/`
- `contact.md` → `/contact/`

---

## Themes

Smolblog has a built-in theme system managed entirely from the admin UI at `/admin/themes`.

- **Install**: upload a `.zip` containing `theme.json`, `templates/`, and optional `static/`
- **Activate**: one click — site rebuilds with the new look
- **Export**: download any installed theme as a `.zip`
- **Delete**: remove themes you don't need (can't delete the active theme or the default)

Each theme is a directory under `themes/<name>/`:

```
my-theme/
├── theme.json           # { "name", "version", "author", "description" }
├── templates/
│   ├── base.html        # defines "base" + "navbar"
│   ├── shared.html      # defines "shared-head"
│   ├── page.html        # defines "content" for static pages
│   ├── post.html        # defines "content" for blog posts
│   └── blog-list.html   # defines "content" for blog listing
└── static/              # optional — copied verbatim to dist/
    └── style.css
```

---

## Project Structure

```
├── .github/workflows/
│   └── deploy.yml         # GitHub Actions deployment
├── content/
│   ├── pages/
│   │   ├── index.md       # Home page source
│   │   ├── about.md       # About page source
│   │   └── contact.md     # Contact page source
│   └── posts/
│       └── *.md           # Blog post sources (YAML frontmatter + Markdown)
├── themes/
│   └── default/
│       ├── theme.json     # Theme metadata
│       ├── templates/     # Go html/template files
│       └── static/        # Optional assets copied to dist/
├── dist/                  # Generated output (deployed to Pages)
├── *.go                   # Go source files
├── index.html             # Admin UI template
├── go.mod / go.sum
└── README.md
```

## Commands

| Command | What it does |
|---|---|
| `go build -o smolblog .` | Build the binary |
| `./smolblog` | Build site + start server (admin + preview) |
| `./smolblog build` | Build site only (for CI/CD) |
| `./smolblog help` | Show usage |
