# CISSP Quiz Trainer

A local-first Windows 11 CISSP practice, review, and learning analytics desktop application.

## Technical baseline

- Go 1.27.x toolchain for development, CI, and release builds
- Wails v2 (current dependency and CLI: v2.15.0)
- Vue 3
- TypeScript
- Pinia
- SQLite
- GitHub Actions

## Documentation

- [Requirements v0.3](docs/requirements-v0.3.md)
- [Codex / contributor instructions](AGENTS.md)

## Repository status

The repository contains the initial Wails + Vue + TypeScript scaffold. Product feature implementation has intentionally not started yet.

## Development prerequisites

- Go 1.27.x toolchain
- Node.js 22.x as the development and CI baseline; other maintained releases require compatibility validation with this project
- npm (no project-specific version is currently pinned)
- Wails CLI v2.15.0, matching the current dependency in `go.mod` and the CI installation command
- WebView2 on Windows

### Version declarations

The Go versions describe different requirements: `go.mod` currently declares `go 1.25.0`, the module's declared minimum Go version and language semantics baseline, while the project selects Go 1.27.x as its development, CI, and release toolchain. The `go` directive does not pin the build toolchain to 1.25.0; see the [Go module reference](https://go.dev/doc/modules/gomod-ref#go). This declaration alone does not establish that the full dependency graph builds with Go 1.25.0.

The current CI configuration selects Go `1.27.x`, Node.js `22`, and Wails CLI `v2.15.0`. These are configured versions, not evidence of a successful build. Node.js 24 or another maintained release is a local alternative only after the required project checks pass; installation alone does not establish compatibility. Frontend dependency ranges are declared in `frontend/package.json`.

Install Wails:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
```

Install frontend dependencies:

```bash
cd frontend
npm install
cd ..
```

Run locally:

```bash
wails dev
```

Validate a change:

```bash
cd frontend
npm run type-check
npm run build
cd ..
gofmt -w .
go vet ./...
go test ./...
wails build -clean
```

## Architecture

```text
Vue / TypeScript
      ↓
Wails binding
      ↓
Go service layer
      ↓
Repository interfaces
      ↓
SQLite
```

See `AGENTS.md` for coding boundaries and `docs/requirements-v0.3.md` for the product baseline.
