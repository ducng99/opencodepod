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

## Testing

```bash
# Run all tests
go test ./internal/ -v -count=1 -timeout 5m
```

- Integration tests use `nginx:alpine` as test image (auto-pulled if missing)
- `skipIfNoDocker()` helper skips the entire Docker suite if daemon is unreachable
- Tests create real containers/volumes and clean them up via `cleanupTestProject()`
- Long timeout recommended: container pulls and start/stop cycles are slow

## Key Conventions

- **Naming**: containers `cp-<id>`, volumes `cp-vol-<id>`. Never look up by name; always by label.
- **Ports**: Docker assigns random host ports for `22/tcp` and `8080/tcp`. Captured via `ContainerInspect` after start.
- **Go 1.26+ routing**: handlers use `http.ServeMux` path patterns like `/api/projects/{id}`
- **Config**: loaded from `config.json` with JSON keys in snake_case (`listen_addr`, `default_image`, etc.). Missing fields fall back to hard-coded defaults.
- **Git auth**: `git.auth.ssh_key` (inline private key) is copied into containers via `CopyToContainer` before start, never as an env var. `git.auth.ssh_key_path` (default `/home/coder/.ssh/id_ed25519`) controls the destination inside the container.

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
- Don't run Tailscale inside containers — it runs only on the host; containers bind `0.0.0.0`
- Don't chase 100% coverage — integration tests cover the Docker lifecycle; unit tests cover parsing and domain logic
