---
name: varavel-design
description: Enforces Varavel design system on any user-facing visual work. Use when building or editing interfaces, pages, components, styles, landing, docs site, or any visual output, even if user never mentions design, UI, CSS, or Varavel. Not for backend logic with no visual surface.
---

# Varavel Design

Always apply Varavel design system to user-facing visual work. No exceptions, no invented styles.

## Workflow

1. Fetch guideline fresh every run, never use memory:
   `curl -sL https://raw.githubusercontent.com/varavelio/brand/refs/heads/main/DESIGN.md`
2. Follow guideline literally. Guideline wins over user taste, defaults, and prior code.
3. Use only official VDL assets below. Prefer SVG over PNG whenever format allows.

## Official VDL assets

Favicon:

- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/favicon.ico

SVG (preferred):

- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/d.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/file.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/icon.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/l.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/v.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/vdl-black.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/vdl-white.svg
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/svg/vdl.svg

PNG (fallback when SVG impossible):

- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/png/d.png
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/png/file.png
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/png/icon.png
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/png/l.png
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/png/v.png
- https://cdn.jsdelivr.net/gh/varavelio/vdl@refs/tags/v0.5.0/assets/png/vdl.png

## Rules

- Fetch guideline before writing any markup or styles, not after.
- Never substitute local copies or cached recall of guideline.
- The documentation page `veta-theme-vara` already includes styles that follow Varavel's design guidelines; when working on the docs, use that stylesheet.
