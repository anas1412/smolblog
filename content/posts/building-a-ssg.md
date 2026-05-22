---
title: Building a Static Site Generator in Go
date: 2026-05-20
slug: building-a-ssg
tags:
  - go
  - tutorial
---

# Building a Static Site Generator in Go

I decided to build my own SSG rather than use Hugo or Jekyll. Here's why.

## Why Build Your Own?

Control. I wanted to understand every part of the pipeline:

1. **Frontmatter parsing** - YAML metadata in markdown files
2. **Markdown rendering** - converting markdown to HTML
3. **Template composition** - base layouts, partials, content blocks
4. **File generation** - writing flat HTML files to disk

## The Stack

- **Language:** Go 1.26
- **Markdown:** Goldmark
- **YAML:** gopkg.in/yaml.v3
- **Templates:** Go `html/template`
- **CMS UI:** HTMX 2.0 + Tailwind CSS

## What I Learned

The hardest part wasn't the code - it was designing the content model. YAML frontmatter gives you metadata without complicating the writing experience.
