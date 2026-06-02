# Agent Notes

This is the isolated Docker workspace for a coding agent. You have full sudo access and network connectivity.

## Environment

- **OS**: Ubuntu LTS latest
- **User**: `ubuntu` (home `/home/ubuntu`)
- **Sudo**: passwordless — `sudo apt-get install ...` works without prompts
- **Shell**: Zsh (with Oh My Zsh)
- **Workspace**: `<root>/workspaces` (owned by `ubuntu`)

## Pre-installed Tools

Some core build/runtime stack:
- `git`, `curl`, `wget`, `build-essential`, `zsh`
- `unzip`, `zip`, `jq`, `htop`, `tree`, `ripgrep`

More tools are available depending on project

## What you can do

- Install any packages with `apt-get` (use `sudo`)
- Create virtualenvs, install pip/npm packages, clone repos, compile code
- Write to `/workspaces` or anywhere writable by `coder`
- Use `sudo` for system-level changes (installing system libs, services, etc.)

## Conventions

- Keep project files under `/workspaces` when possible
