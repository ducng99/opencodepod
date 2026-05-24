# OpenCodePod

[![Build & Release](https://github.com/ducng99/opencodepod/actions/workflows/release.yml/badge.svg)](https://github.com/ducng99/opencodepod/actions/workflows/release.yml)
[![Tests](https://github.com/ducng99/opencodepod/actions/workflows/test.yml/badge.svg)](https://github.com/ducng99/opencodepod/actions/workflows/test.yml)

OpenCodePod is a lightweight, stateless Go server that turns Docker containers into isolated project workspaces. It manages the full lifecycle of development environments — create, start, stop, and delete — with zero database dependency. All state lives in Docker container labels and live container queries.

## Features

- **Project Workspaces** — Each project gets its own Docker container and persistent volume.
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

`"{file:./git.key}"` syntax can be used to inline the content of the file into a field. See config example below how it is used for GPG key.<br/>
Note: The content will be trimmed for both prefix and suffix.

| JSON key | Type | Default | Description |
|----------|------|---------|-------------|
| `listen_addr` | string | `127.0.0.1:8080` | HTTP listen address. Listening on localhost only, if you want it accessible globally, or running in docker, use `0.0.0.0:8080` |
| `default_image` | string | `ghcr.io/ducng99/opencodepod-client:latest` | Default Docker image for new projects |
| `ssh_public_key` | string | *(empty)* | SSH public key injected into containers via `SSH_PUBLIC_KEY` env |
| `mounts` | array | `[]` | Extra host → container mounts applied to every project. Each item has `source` (host path), `target` (container path), and optional `read_only` (boolean). |
| `git.auth.ssh_key` | string | *(empty)* | Inline SSH private key used for cloning private git repositories. Copied into containers before start via the Docker API. |
| `git.auth.ssh_key_path` | string | `/home/coder/.ssh/id_ed25519` | Destination path inside the container where the SSH key is copied. |
| `git.auth.credentials` | object | `{}` | Host-keyed username/password credentials for Git HTTP authentication. Each key is a hostname (e.g. `github.com`) and each value has `username` and `password` fields. Copied into containers as `~/.git-credentials` with `credential.helper store` configured automatically. |
| `hosts` | object | `{}` | Custom host entries added to container `/etc/hosts`. |
| `git.user_name` | string | *(empty)* | Git commit author name. |
| `git.user_email` | string | *(empty)* | Git commit author email. |
| `git.gpg.key_id` | string | *(empty)* | GPG key ID used for commit signing. |
| `git.gpg.private_key` | string | *(empty)* | Inline GPG private key for signing commits. |
| `git.gpg.passphrase` | string | *(empty)* | Passphrase for the GPG private key. Copied into containers so `git commit -S` can run non-interactively. |

### Example `config.json`

```json
{
  "$schema": "https://raw.githubusercontent.com/ducng99/opencodepod/refs/heads/main/config.schema.json",
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
  "hosts": {
    "my-registry.local": "10.0.0.5"
  },
  "git": {
    "user_name": "My Name",
    "user_email": "my@email.com",
    "auth": {
      "ssh_key": "{file:<host_path_to_ssh_key>}",
      "ssh_key_path": "/home/coder/.ssh/id_ed25519",
      "credentials": {
        "github.com": {
          "username": "myuser",
          "password": "github_pat_xxx"
        },
        "gitlab.company.internal": {
          "username": "me",
          "password": "idk"
        }
      }
    },
    "gpg": {
      "key_id": "A1B2C3D4E5F6",
      "private_key": "{file:<host_path_to_gpg_key>}",
      "passphrase": "{file:<host_path_to_gpg_passphrase>}"
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

**SSH key authentication**

Set `git.auth.ssh_key` in `config.json` to your SSH private key. It is copied into containers automatically before startup. The default path inside the container is `/home/coder/.ssh/id_ed25519`; change it with `git.auth.ssh_key_path` if needed.

> [!NOTE]
> SSH keys with a passphrase are not supported because there is no interactive prompt to unlock them during container startup.
> Either clone yourself after starting the project, or use HTTPS with a fine-grained personal access token (PAT).

**HTTP credentials**

For HTTPS repositories, set `git.auth.credentials` with host-keyed username/password (PAT) pairs:

```json
"git": {
  "auth": {
    "credentials": {
      "github.com": {
        "username": "myuser",
        "password": "github_ghp_xxx" // or "{file:/app/github_pat.key}" to inline the file content
      }
    }
  }
}
```

Git credential helper is configured automatically so clones proceed without prompts.

### GPG signing

To sign commits inside containers, set `git.gpg.key_id` and `git.gpg.private_key` in `config.json`.

You can export your GPG with

```sh
gpg --armor --export-secret-keys "YOUR_KEY_ID" > gpg.key
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
- **Naming** — Containers are named `cp-<id>`, volumes `cp-vol-<id>`.
- **Port assignment** — Docker randomly assigns host ports for `22/tcp` and `8080/tcp`. CodePod inspects the container after start to discover them.
- **Volumes** — Each project gets a dedicated Docker volume mounted at `/workspaces` inside the container.
- **Restart policy** — Unless stopped via the API, containers use Docker's `unless-stopped` restart policy.

## Development

### Building

```bash
# Build frontend (requires Bun)
cd frontend && bun install && bun run build && cd ..

# Build Go server
go build -o opencodepod-server ./cmd/server

# Build client Docker image
docker build -t opencodepod-client:latest -f docker/Dockerfile ./docker
```

### Testing

```bash
# Run all tests
go test ./... -v -count=1 -parallel 4 -timeout 5m
```

Integration tests use `nginx:alpine` as a test image and create real containers/volumes, cleaning them up afterward. Most tests are marked `t.Parallel()` for concurrent execution.

## License

MIT
