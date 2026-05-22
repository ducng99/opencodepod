---
description: git commit
subtask: true
---

commit and push

use conventional commit prefixes:
- feat: - new feature
- fix: - bug fix
- docs: - documentation only
- style: - formatting, missing semicolons, etc
- refactor: - code change that neither fixes a bug nor adds a feature
- perf: - performance improvement
- test: - adding or correcting tests
- build: - build system or dependencies
- ci: - CI configuration
- chore: - maintenance tasks
- revert: - reverting a previous commit

use the docs: prefix for changes to frontend/ or packages/web

prefer to explain WHY something was done from an end user perspective instead of
WHAT was done.

do not do generic messages like "improved agent experience" be very specific
about what user facing changes were made

if there are conflicts DO NOT FIX THEM. notify me and I will fix them

## GIT DIFF

!`git diff`

## GIT DIFF --cached

!`git diff --cached`

## GIT STATUS --short

!`git status --short`
