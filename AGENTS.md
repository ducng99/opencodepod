# OpenCodePod — Agent Notes

Stateless Go orchestrator that manages Docker containers as isolated project workspaces. No database — all runtime state comes from Docker labels and live container queries.

## Architecture

- **Entry**: `cmd/server/main.go` — boots HTTP server on `listen_addr` (default `:8080`)
- **Packages**:
    - `internal/` — all business logic (handlers, docker client, config, domain types)
    - `frontend/` — static UI served via `//go:embed all:dist`
- **State**: Docker container labels (`opencodepod.managed=true`, `opencodepod.project.id`, etc.) are the source of truth. Server restart = re-list containers.

## Build & Run

```bash
# Build frontend first (requires Bun)
cd frontend && bun install && bun run build && cd ..

# Build Go server
go build -o opencodepod-server ./cmd/server
./opencodepod-server
```

No Makefile or task runner. Dockerfile is a standard multi-stage Alpine build.

**Frontend build details**: Custom `frontend/build.ts` script — `bun build` bundles JS/TSX, `tailwindcss` CLI compiles CSS, then copies `index.html` into `dist/`.

## Testing

```bash
# Run all tests (long timeout recommended — container pulls are slow)
go test ./... -v -count=1 -parallel 4 -timeout 5m

# Run specific tests (by func name partial match)
go test ./... -v -count=1 -parallel 4 -timeout 5m -run 'TestConfig|Test...'
```

- Integration tests use `nginx:alpine` as test image (auto-pulled if missing)
- Tests create real containers/volumes and clean them up via `cleanupTestProject()`
- Most tests are marked `t.Parallel()` for concurrent execution; adjust `-parallel` to match your CPU cores
- CI order: build frontend → `go vet ./...` → test

## Key Conventions

- **Naming**: containers `cp-<id>`, volumes `cp-vol-<id>` and `cp-vol-<id>-home`. Never look up by name; always by label.
- **Ports**: Docker assigns random host ports for `22/tcp` and `8080/tcp`. Captured via `ContainerInspect` after start.
- **Go 1.26+ routing**: handlers use `http.ServeMux` path patterns like `/api/projects/{id}`
- **Config**: loaded from `config.json` with JSON keys in snake_case (`listen_addr`, `default_image`, etc.). Missing fields fall back to hard-coded defaults.
- **File placeholders**: config fields support `{file:<host_path>}` syntax — the file content is inlined at load time. Relative paths resolve against `config.json`'s directory.
- **Schema**: whenever `internal/config/config.go` structs or their `desc` tags change, regenerate `config.schema.json` with `go run ./cmd/generate-schema`.
- **Git auth**: `git.auth.ssh_key` (inline private key or `{file:...}`) is copied into containers via `CopyToContainer` before start, never as an env var. `git.auth.ssh_key_path` (default `/home/coder/.ssh/id_ed25519`) controls the destination inside the container.

## Frontend

React 19 + TypeScript + Tailwind CSS v4, built with Bun.

- **Source**: `frontend/src/` — components, API client, types
- **Build output**: `frontend/dist/` — served via `//go:embed all:dist`
- **Build step required**: `cd frontend && bun install && bun run build`
- Auto-regenerates `dist/index.html`, `dist/main.js`, `dist/index.css`
- `frontend/embed.go` includes `//go:generate bun run build` for convenience

### Tech stack

- React 19 with `useState`/`useEffect` (no external state library)
- Tailwind CSS v4 with CSS-based configuration (`@import "tailwindcss"`)
- Bun as package manager and bundler (`bun build` for JS/TSX, `tailwindcss` CLI for CSS)
- Polling: `GET /api/projects` every 5s via `setInterval`

## What NOT to do

- Don't add a database for project state — query Docker live on every request
- Don't expose the Docker socket over TCP — mount read-only UNIX socket or skip Docker calls entirely
- Don't chase 100% coverage — integration tests cover the Docker lifecycle; unit tests cover parsing and domain logic
