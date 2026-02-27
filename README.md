# t4gfm

`t4gfm` means **time for good file manager**.

This project is an opinionated terminal file manager tuned for a ranger-like keyboard workflow, safer defaults, and practical terminal-first UX.

## Project Attribution

- Repository: https://github.com/t4Linux/t4gfm
- License: see [LICENSE](./LICENSE)

## What This Fork Focuses On

- ranger-style key workflow and prefix actions
- custom sidebar + compact/footer workflow
- safer open behavior for risky/binary/large files
- symlink-aware navigation and preview behavior
- performance improvements for day-to-day use

## Installation

Install from this fork:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/t4Linux/t4gfm/main/install.sh)"
```

## Build

Requirements:

- Go (latest stable)

Build locally:

```bash
git clone https://github.com/t4Linux/t4gfm.git --depth=1
cd t4gfm
go build -o bin/gfm .
```

Run:

```bash
./bin/gfm
```

Or install to PATH:

```bash
sudo mv ./bin/gfm /usr/local/bin/gfm
```

## Project Docs

- Changelog: [CHANGELOG.md](./CHANGELOG.md)
- Release notes: [RELEASE_NOTES_v0.1.0.md](./RELEASE_NOTES_v0.1.0.md)

## Shell Integration (cd on quit)

To make terminal directory follow the last focused directory after quit, load the shell wrapper.

```bash
source /path/to/t4gfm/cd_on_quit/cd_on_quit.sh
```

Fish:

```fish
source /path/to/t4gfm/cd_on_quit/cd_on_quit.fish
```

## Theme Switching

- One-off run with theme override: `gfm --theme gruvbox-dark-hard`
- Back to default theme for one run: `gfm --theme catppuccin-mocha`
- Optional toggle helpers (bash/zsh):

```bash
source /path/to/t4gfm/cd_on_quit/theme_toggle.sh
gfm_toggle      # switches between default and gruvbox-dark-hard
gfm_default     # force default
gfm_gruvbox     # force gruvbox-dark-hard
```

## Full Keyboard Shortcuts

### Global

| Action | Keys |
|---|---|
| Confirm/open | `enter`, `right`, `l` |
| Quit (exports last directory for shell wrapper) | `q`, `Q` |
| Move up/down | `up`/`k`, `down`/`j` |
| Page up/down | `ctrl+b`/`pgup`, `ctrl+f`/`pgdown` |
| Close/create panel | `w`, `n` |
| Next/previous panel | `tab`, `shift+left`/`H` |
| Pin current directory | `P` |
| Open sort options menu | `O` |
| Toggle preview panel | `f` |
| Toggle reverse sort | `ctrl+r` |
| Move pinned entry | `K`/`shift+k` up, `J`/`shift+j` down |
| Create/rename item | `a`, `r` |
| Copy/cut/paste | `y`, `x` |
| Two-panel transfer | `C` (copy to other panel), `M` (move to other panel) |
| Delete/trash | `delete` |
| Permanent delete | `D` |
| Open dir/file with editor | `E`, `R`/`e` |
| Toggle select mode | `space`, `v` |
| Copy path | `Y` |
| Command line / prompt / zoxide | `:`, `>`, `Z` |
| Help | `?` |
| Toggle dotfiles/sidebar/footer | `.`, `b`, `F` |

### Typing Mode

| Action | Keys |
|---|---|
| Confirm typing | `enter` |
| Cancel typing | `ctrl+c`, `esc` |

### Mode-specific

| Mode | Action | Keys |
|---|---|---|
| Normal | Parent directory | `h`, `left` |
| Normal | Previous location (`cd -` style) | `backspace` |
| Normal | Search bar | `/` |
| Selection | Expand selection down/up | `shift+down`/`J`, `shift+up`/`K` |
| Selection | Select all | `A` |

### Ranger-style Prefix Menus

Press prefix key first, then second key(s):

- `y` copy submenu: `yy`, `yp`, `yd`, `yn`, `y.`
- `d` cut/delete submenu: `dd`, `dT`, `dD`, `du`
- `p` paste submenu: `pp`, `po`, `pP`, `pO`, `pl`, `pL`, `phl`, `pht`
- `g` go submenu: `gg`, `gh`, `gl`, `g/`, `gp`, `ge`, `gu`, `gd`, `go`, `gv`, `gm`, `gM`, `gs`
- `o` sort submenu: `or`, `on`, `os`, `om`, `ot`, `oN`, `oS`, `oM`, `oT`
- `z` toggle/search submenu: `zh`, `zf`
- `s` focus/jump submenu: `ss`, `sm`, `sa`, `sg`, `sp`, `sd`, `sf`
- `m` save mark submenu: `mxy` (two-letter mark key)
- `;` marks submenu: `;?`, `;xy`, `;dxy`, `;D`
- `c` compress submenu:
  - `cz` `.zip`
  - `ct` `.tar.gz`
  - `cj` `.tar.xz`
  - `cl` `.tar.zst`
  - `cE` encrypted `.tar.gz.gpg` (passphrase prompt)
  - `cs` cycle compression level
  - `cv` toggle verbose tar mode
  - `ce` toggle common excludes
- `+`/`-` chmod submenu:
  - quick: `+r/+w/+x/+X/+s/+t` and `-r/-w/-x/-X/-s/-t`
  - class-first: `+u/+g/+o/+a` then `r/w/x/X/s/t` (same for `-`)

### Panel-specific Extras

- Git panel tabs: `H` and `L`
- File preview scrolling: `Shift+J`, `Shift+K`
