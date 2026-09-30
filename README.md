# CISSP Quiz Trainer

A local-first Windows 11 CISSP practice, review, and learning analytics desktop application.

## Technical baseline

- Go 1.27.x toolchain
- Wails v2.15.x
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

- Go 1.27.x
- Node.js 22.x
- npm
- Wails CLI v2.15.x
- WebView2 on Windows

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
