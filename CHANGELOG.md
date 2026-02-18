# Changelog

All notable changes in this fork are documented in this file.

Project name: **t4gfm** (time for good file manager)

## [0.1.0] - 2026-02-18

### Added
- Ranger-inspired key workflow with prefix menus (`y*`, `d*`, `p*`, `g*`, `s*`) and extended mark navigation.
- New footer mode system with compact one-line footer containing file info, git summary, and right-side disk/progress info.
- Sidebar section navigation shortcuts (`sp`, `sd`, `sl`) and multi-panel sidebar layout (`List`, `Pinned`, `Disks`).
- Symlink-aware navigation controls including physical path jump (`gl`).
- Visual selection mode via `V`/`shift+v`, with range extension and `o` endpoint swap like nvim visual line behavior.
- Fork installer script (`install.sh`) and fork-specific setup instructions.

### Changed
- Replaced legacy footer panel usage with customized `System`, `Metadata`, and `Git` behavior.
- Improved copy progress reporting and compact process display in footer.
- Updated color and highlight behavior for selected items and panel focus states.
- Updated metadata presentation for path visibility and symlink-specific details.
- Revised default hotkey behavior around panel focus and ranger-like interactions.

### Fixed
- Restored and stabilized dynamic tmux title behavior in custom workflows.
- Fixed multiple sidebar/runtime regressions around toggle, mouse focus, and section jumps.
- Fixed symlink parent navigation behavior (logical vs physical path backtracking).
- Fixed extensionless text preview handling and symlink-target preview fallback.
- Fixed cancellation reliability for long-running copy operations.

### Performance
- Added caching/throttling for directory entry counts in file list size column.
- Added throttled git marker refresh to avoid excessive status calls.
- Added sidebar refresh throttling and reduced repeated expensive refresh paths.
- Added metadata fetch debounce and background panel sync throttling.

### Security / Safety
- Added open guards for unsafe file opens (binary detection, size limit, blocked VM image extensions).
- Unified open-block policy across metadata “Open with” and actual open commands.
