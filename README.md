# OpenCodePod

OpenCodePod is a lightweight, stateless Go server that turns Docker containers into isolated project workspaces. It manages the full lifecycle of development environments — create, start, stop, and delete — with zero database dependency. All state lives in Docker container labels and live container queries.

## Features

- **Project Workspaces** — Each project gets its own isolated Docker container.
- **Web UI** — Dark-themed dashboard at `/` for managing projects without touching the CLI.
- **REST API** — Full JSON API for automation and integrations.
- **SSH & Web Access** — Containers expose `22/tcp` (SSH) and `8080/tcp` (web app); Docker assigns random host ports automatically.
- **Git Clone on Boot** — Pass a `git_repo` and the container receives it via the `GIT_REPO` environment variable.
- **SSH Public Key Injection** — Configure a global SSH public key in `config.json`.
- **Extra Mounts** — Mount host files or directories into every project container via JSON config.
- **Host Gateway Access** — Every container gets `host.docker.internal` mapped to the host gateway, enabling access to host services.
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
./opencodepod-server
```

The server starts on `http://localhost:8080` by default. Open your browser to `http://localhost:8080` to use the web UI.

## Configuration

Configuration is loaded from `config.json` in the working directory. Missing fields fall back to hard-coded defaults.

| JSON key | Type | Default | Description |
|----------|------|---------|-------------|
| `listen_addr` | string | `:8080` | HTTP listen address |
| `default_image` | string | `ghcr.io/ducng99/opencodepod-client:latest` | Default Docker image for new projects |
| `ssh_public_key` | string | *(empty)* | SSH public key injected into containers via `SSH_PUBLIC_KEY` env |
| `mounts` | array | `[]` | Extra host → container mounts applied to every project. Each item has `source` (host path), `target` (container path), and optional `read_only` (boolean). |
| `git.auth.ssh_key` | string | *(empty)* | Inline SSH private key used for cloning private git repositories. Copied into containers before start via the Docker API. |
| `git.auth.ssh_key_path` | string | `/home/coder/.ssh/id_ed25519` | Destination path inside the container where the SSH key is copied. |

### Example `config.json`

```json
{
  "listen_addr": "0.0.0.0:8080",
  "default_image": "ghcr.io/ducng99/opencodepod-client:latest",
  "ssh_public_key": "ssh-ed25519 AAAAC3NzaC...",
  "mounts": [
    {
      "source": "/host/path/to/opencode.jsonc",
      "target": "/home/coder/.config/opencode/opencode.jsonc",
      "read_only": true
    }
  ],
  "git": {
    "auth": {
      "ssh_key": "{file:<host_path_to_ssh_key>}",
      "ssh_key_path": "/home/coder/.ssh/id_ed25519"
    }
  }
}
```

### Docker Compose example

```yaml
services:
  server:
    image: ghcr.io/ducng99/opencodepod-server:latest
    ports:
      - 10000:8080
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./config.json:/app/config.json:ro
```

### Cloning private repositories

Set `git.auth.ssh_key` in `config.json` to an inline SSH private key. The server copies it into each new container before startup (never via env vars). The default destination is `/home/coder/.ssh/id_ed25519`; you can override it with `git.auth.ssh_key_path`.

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

`git_repo` and `image` are optional. If `image` is omitted, the configured `default_image` is used.

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

Deletes the container. This cannot be undone.

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
  -v $(pwd)/config.json:/app/config.json:ro \
  --name opencodepod \
  opencodepod
```

> **Note:** Mount the Docker socket read-only. OpenCodePod needs to create and manage containers on your behalf.

## How It Works

- **No database** — Project state is stored entirely in Docker labels (`opencodepod.managed=true`, `opencodepod.project.id`, etc.).
- **Naming** — Containers are named `cp-<id>`.
- **Port assignment** — Docker randomly assigns host ports for `22/tcp` and `8080/tcp`. CodePod inspects the container after start to discover them.
- **Restart policy** — Unless stopped via the API, containers use Docker's `unless-stopped` restart policy.

## Development

### Project Structure

```
opencodepod/
├── cmd/server/main.go      # Entry point
├── internal/
│   ├── config.go           # JSON config loader
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
# Run all tests
go test ./internal/ -v -count=1 -timeout 5m
```

Integration tests use `nginx:alpine` as a test image and create real containers, cleaning them up afterward.

## License

MIT
