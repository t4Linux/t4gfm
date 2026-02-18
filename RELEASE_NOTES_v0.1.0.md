# Release Notes - v0.1.0

Project name: **t4gfm** (time for good file manager)

## Summary

`v0.1.0` is the first fork-focused release with a strong ranger-like workflow, custom panel behavior, safer file opening defaults, and performance-oriented runtime tuning.

## Highlights

- Ranger-style interaction model with extended prefix menus and mark workflows.
- Sidebar redesigned into independent `List`, `Pinned`, and `Disks` sections.
- New compact footer mode with one-line operational context.
- Symlink workflows improved for both navigation and preview consistency.
- Visual range selection (`V`) with directional endpoint switching (`o`) like nvim.
- Runtime optimizations for metadata, git markers, sidebar refresh, and file list rendering.
- Safer open behavior for large/binary/VM image files.

## Breaking/Behavioral Changes

- Footer toggle behavior now supports compact/full modes in fork workflow.
- Some hotkeys were remapped for ranger-like consistency and panel focus control.
- Unsafe/binary large files may now be blocked from opening by default.

## Upgrade Notes

1. Rebuild or reinstall the fork binary.
2. Review `~/.config/t4gfm/hotkeys.toml` for fork-aligned key mappings.
3. If migrating from upstream defaults, run with `--fix-hotkeys` once.

## Attribution

This release is based on the upstream project:
- https://github.com/t4Linux/t4gfm

The fork keeps upstream licensing and attribution.
