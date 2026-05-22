---
title: Why Go, HTMX, and Tailwind Are a Great Pair
date: 2026-05-18
slug: htmx-tailwind-love
tags:
  - go
  - htmx
  - tailwind
  - architecture
---

# Why Go, HTMX, and Tailwind Are a Great Pair

Smolblog is built with three tools that, on paper, seem unrelated - Go, HTMX, and Tailwind CSS. But together they form something genuinely nice: a stack that's simple, fast, and fun to work with.

## Go - the solid foundation

Go compiles to a single binary. No runtime, no VM, no dependency hell. For Smolblog, that means:

- **Instant startup** - the server is ready in milliseconds
- **Sub-second builds** - even with dozens of posts
- **One dependency** - everything in `go.mod` fits on one screen

Go's simplicity means you can understand the entire codebase in an afternoon. That's rare and valuable.

## HTMX - interactivity without complexity

The admin panel needs to be interactive - save posts, build the site, switch themes. HTMX handles all of that with HTML attributes, not a frontend framework:

```html
<button hx-post="/admin/build" hx-target="#status">
    Build Site
</button>
```

That's it. No fetch calls, no state management, no bundles. HTMX sends the request and swaps the response into the target element. The server returns HTML, not JSON. It's how the web worked before we overcomplicated it.

## Tailwind - styling that stays out of the way

Tailwind keeps the CSS in the HTML where it belongs. No separate stylesheets to hunt through, no naming conventions to maintain:

```html
<nav class="border-b border-zinc-100 bg-white/80 backdrop-blur sticky top-0">
```

Every class is a constraint, not a convention. You build UI by composing utilities, and the result is consistent without a design system document.

## The whole is greater than the sum

The magic is what these three don't do:

- **No JavaScript framework** to learn, build, or debug
- **No CSS preprocessor** or design token pipeline
- **No runtime dependencies** to track or update

Just a compiled Go binary, HTML attributes for interactivity, and utility classes for style. It's the most productive stack I've used in years.
