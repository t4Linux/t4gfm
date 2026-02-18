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
