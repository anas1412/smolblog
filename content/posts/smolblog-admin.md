---
title: Managing Smolblog Without Touching Code
date: 2026-05-22
slug: smolblog-admin
tags:
  - smolblog
  - cms
  - themes
---

# Managing Smolblog Without Touching Code

One of the best things about Smolblog is that once it's set up, you never need to touch a terminal or edit a file directly. Everything is managed from the admin panel.

## Writing Posts

The admin panel at `/admin` gives you a clean editor with:

- **Title, slug, date, and tags** - all the metadata you need
- **Markdown editor** with live preview (via EasyMDE)
- **Save** button that writes the file and rebuilds the site instantly

No need to remember YAML frontmatter syntax or file naming conventions. The admin handles all of that.

## Settings

The Settings page lets you change:

- **Site title** - shown in the navbar and browser tabs
- **Footer text** - displayed at the bottom of every page
- **Home and About page content** - written in Markdown, rendered live

All saved through a simple form. No config files to edit.

## Theme Management

The Themes page is where Smolblog really shines:

- **Browse installed themes** - see what's available at a glance
- **Install new themes** - upload a `.zip` file and it's ready to use
- **Activate** - one click switches your site's entire look
- **Export** - download any theme to share or back up
- **Delete** - remove themes you don't need

Themes are just templates and static assets packaged in a zip. You can design your own or use community themes. No code changes, no rebuilding the binary.

## Build & Deploy

- **Build** button regenerates the static site
- **Publish** button commits and pushes to GitHub (requires git + SSH setup)

That's it. A fully functional blog you manage entirely from the browser.
