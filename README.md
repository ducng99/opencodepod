# OpenCodePod

OpenCodePod is a lightweight, stateless Go server that turns Docker containers into isolated project workspaces. It manages the full lifecycle of development environments — create, start, stop, and delete — with zero database dependency. All state lives in Docker container labels and live container queries.

## Features

- **Project Workspaces** — Each project gets its own Docker container and persistent volume.
- **Web UI** — Dark-themed dashboard at `/` for managing projects without touching the CLI.
- **REST API** — Full JSON API for automation and integrations.
- **SSH & Web Access** — Containers expose `22/tcp` (SSH) and `8080/tcp` (web app); Docker assigns random host ports automatically.
- **Git Clone on Boot** — Pass a `git_repo` and the container receives it via the `GIT_REPO` environment variable.
- **SSH Public Key Injection** — Configure a global SSH public key via environment variable.
- **Resource Limits** — Optionally set CPU and memory constraints for all project containers.
- **Stateless** — Server restart? No problem. State is reconstructed by querying Docker on every request.

## Prerequisites

- [Go](https://go.dev/) 1.26+
- [Docker](https://docs.docker.com/get-docker/) Engine (running locally or accessible via socket)

## Quick Start

### Build

```bash
go build -o opencodepod-server ./cmd/server
```

### Run

```bash
./codepod-server
```

The server starts on `http://localhost:8080` by default. Open your browser to `http://localhost:8080` to use the web UI.

## Configuration

All configuration is environment-driven:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_LISTEN` | `:8080` | HTTP listen address |
| `DEFAULT_IMAGE` | `ghcr.io/ducng99/opencodepod-client:latest` | Default Docker image for new projects |
| `APP_SSH_PUBLIC_KEY` | *(empty)* | SSH public key injected into containers via `SSH_PUBLIC_KEY` env |

### Example with custom config

```bash
export APP_LISTEN=:3000
export DEFAULT_IMAGE=ghcr.io/ducng99/opencodepod-client:latest
export APP_SSH_PUBLIC_KEY="ssh-ed25519 AAAAC3NzaC..."
./opencodepod-server
```

## Using the Web UI

1. Open `http://localhost:8080`
2. Enter a **Project name** (required)
3. Optionally provide a **Git repo** URL and a custom **Docker image**
4. Click **+ New Project**
5. The grid shows your project with:
   - Status badge (Running / Stopped / Creating)
   - SSH command (`ssh -p <port> coder@<host>`)
   - Web URL (`http://<host>:<port>`)
   - Start / Stop / Delete actions

The UI polls the API every 5 seconds to keep the status grid fresh.

## REST API

Base URL: `http://localhost:8080/api`

### List all projects

```http
GET /api/projects
```

Response:
```json
[
  {
    "id": "a1b2c3d4",
    "name": "My Project",
    "git_repo": "https://github.com/user/repo.git",
    "status": "running",
    "ssh_port": 49152,
    "web_port": 49153,
    "volume": "cp-vol-a1b2c3d4",
    "image": "ghcr.io/ducng99/opencodepod-client:latest"
  }
]
```

### Create a project

```http
POST /api/projects
Content-Type: application/json

{
  "name": "My Project",
  "git_repo": "https://github.com/user/repo.git",
  "image": "ghcr.io/ducng99/opencodepod-client:latest"
}
```

`git_repo` and `image` are optional. If `image` is omitted, `DEFAULT_IMAGE` is used.

### Get a project

```http
GET /api/projects/{id}
```

### Start a project

```http
POST /api/projects/{id}/start
```

### Stop a project

```http
POST /api/projects/{id}/stop
```

### Delete a project

```http
DELETE /api/projects/{id}
```

Deletes the container **and** its persistent volume. This cannot be undone.

### Refresh ports

```http
GET /api/projects/{id}/ports
```

Re-inspects the container and returns the project with refreshed port mappings.

## Docker Deployment

You can also run OpenCodePod itself inside Docker:

```bash
docker build -t opencodepod .
docker run -d \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e DEFAULT_IMAGE=ghcr.io/ducng99/opencodepod-client:latest \
  --name opencodepod \
  opencodepod
```

> **Note:** Mount the Docker socket read-only. OpenCodePod needs to create and manage containers on your behalf.

## How It Works

- **No database** — Project state is stored entirely in Docker labels (`opencodepod.managed=true`, `opencodepod.project.id`, etc.).
- **Naming** — Containers are named `cp-<id>`, volumes `cp-vol-<id>`.
- **Port assignment** — Docker randomly assigns host ports for `22/tcp` and `8080/tcp`. CodePod inspects the container after start to discover them.
- **Volumes** — Each project gets a dedicated Docker volume mounted at `/home/coder/workspace` inside the container.
- **Restart policy** — Unless stopped via the API, containers use Docker's `unless-stopped` restart policy.

## Development

### Project Structure

```
opencodepod/
├── cmd/server/main.go      # Entry point
├── internal/
│   ├── config.go           # Environment config
│   ├── docker.go           # Docker lifecycle manager
│   ├── handlers.go         # HTTP handlers
│   ├── project.go          # Domain types & label helpers
│   └── *_test.go           # Unit & integration tests
├── frontend/
│   ├── embed.go            # Go embed for static files
│   └── dist/index.html     # Single-file vanilla JS UI
├── Dockerfile
├── go.mod
└── AGENTS.md               # Agent/developer notes
```

### Testing

```bash
# Run all tests (includes Docker integration tests — needs Docker daemon)
go test ./internal/ -v -count=1 -timeout 5m

# Run only unit tests (no Docker needed)
go test ./internal/ -v -count=1 -run 'TestLabels|TestProject|TestVolume|TestContainer|TestParse|TestUnits|TestConfig|TestGet|TestHandleCreate_BadRequest'
```

Integration tests use `nginx:alpine` as a test image and create real containers/volumes, cleaning them up afterward.

## License

MIT
