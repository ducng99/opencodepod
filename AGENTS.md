# CodePod — Agent Notes

Stateless Go orchestrator that manages Docker containers as isolated project workspaces. No database — all runtime state comes from Docker labels and live container queries.

## Architecture

- **Entry**: `cmd/server/main.go` — boots HTTP server on `APP_LISTEN` (default `:8080`)
- **Packages**:
  - `internal/` — all business logic (handlers, docker client, config, domain types)
  - `frontend/` — static UI served via `//go:embed all:dist`
- **State**: Docker container labels (`codepod.managed=true`, `codepod.project.id`, etc.) are the source of truth. Server restart = re-list containers.

## Build & Run

```bash
go build -o codepod-server ./cmd/server
./codepod-server
```

No Makefile or task runner. Dockerfile is a standard multi-stage Alpine build.

## Testing

```bash
# Run all tests (includes Docker integration tests — need Docker daemon)
go test ./internal/ -v -count=1 -timeout 5m

# Run only unit tests (no Docker needed)
go test ./internal/ -v -count=1 -run 'TestLabels|TestProject|TestVolume|TestContainer|TestParse|TestUnits|TestConfig|TestGet|TestHandleCreate_BadRequest'
```

- Integration tests use `nginx:alpine` as test image (auto-pulled if missing)
- `skipIfNoDocker()` helper skips the entire Docker suite if daemon is unreachable
- Tests create real containers/volumes and clean them up via `cleanupTestProject()`
- Long timeout recommended: container pulls and start/stop cycles are slow

## Key Conventions

- **Naming**: containers `cp-<id>`, volumes `cp-vol-<id>`. Never look up by name; always by label.
- **Ports**: Docker assigns random host ports for `22/tcp` and `8080/tcp`. Captured via `ContainerInspect` after start.
- **Go 1.26+ routing**: handlers use `http.ServeMux` path patterns like `/api/projects/{id}`
- **Config**: all env-driven (`APP_LISTEN`, `APP_TAILNET_HOST`, `DEFAULT_IMAGE`, `APP_SSH_PUBLIC_KEY`)

## Frontend

Single vanilla-JS HTML file in `frontend/dist/index.html`. Dark-themed, polls `/api/projects` every 5s. No build step — edit the file directly.

## What NOT to do

- Don't add a database for project state — query Docker live on every request
- Don't expose the Docker socket over TCP — mount read-only UNIX socket or skip Docker calls entirely
- Don't run Tailscale inside containers — it runs only on the host; containers bind `0.0.0.0`
- Don't chase 100% coverage — integration tests cover the Docker lifecycle; unit tests cover parsing and domain logic
