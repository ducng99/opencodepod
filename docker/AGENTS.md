# CodePod Container — Agent Notes

This is the isolated Docker workspace for a coding agent. You have full sudo access and network connectivity.

## Environment

- **OS**: Ubuntu 24.04 (minimal server image)
- **User**: `coder` (UID auto-assigned, home `/home/coder`)
- **Sudo**: passwordless — `sudo apt-get install ...` works without prompts
- **Shell**: Bash
- **Workspace**: `/workspace` (owned by `coder`)

## Pre-installed Tools

Core build/runtime stack:
- `git`, `curl`, `wget`, `build-essential`
- `python3`, `python3-pip`, `python3-venv`
- `nodejs` (via apt — may be an older LTS; upgrade with `npm install -g n` if you need a specific version)
- `vim`, `nano`, `unzip`, `zip`, `jq`, `htop`, `tree`
- `openssh-server` (sshd runs on container start; public key auth only)

## What you can do

- Install any packages with `apt-get` (use `sudo`)
- Create virtualenvs, install pip/npm packages, clone repos, compile code
- Write to `/workspace` or anywhere writable by `coder`
- Use `sudo` for system-level changes (installing system libs, services, etc.)

## Conventions

- Keep project files under `/workspace` when possible
- If you need a specific language version (e.g., newer Node or Python), install it inside the container rather than modifying the base image
- The host binds random external ports for SSH and the web UI; you don't need to know them unless you are debugging networking
- SSH access is configured automatically if the host sets `SSH_PUBLIC_KEY` — the key is written to `~coder/.ssh/authorized_keys` on container start
