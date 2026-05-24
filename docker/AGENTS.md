# Agent Notes

This is the isolated Docker workspace for a coding agent. You have full sudo access and network connectivity.

## Environment

- **OS**: Ubuntu LTS latest
- **User**: `coder` (UID auto-assigned, home `/home/coder`)
- **Sudo**: passwordless — `sudo apt-get install ...` works without prompts
- **Shell**: Bash
- **Workspace**: `/workspaces` (owned by `coder`)

## Pre-installed Tools

Some core build/runtime stack:
- `git`, `curl`, `wget`, `build-essential`
- `uv` (Astral's Python toolchain)
- `nodejs` (via apt — may be an older LTS)
- `unzip`, `zip`, `jq`, `htop`, `tree`, `ripgrep`

More tools are available.

## What you can do

- Install any packages with `apt-get` (use `sudo`)
- Create virtualenvs, install pip/npm packages, clone repos, compile code
- Write to `/workspaces` or anywhere writable by `coder`
- Use `sudo` for system-level changes (installing system libs, services, etc.)

## Conventions

- Keep project files under `/workspaces` when possible
