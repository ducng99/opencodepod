#!/bin/sh
set -e

WORKSPACE="/home/coder/workspace"

if [ -n "$GIT_REPO" ]; then
  echo "Cloning $GIT_REPO into $WORKSPACE ..."
  cd "$WORKSPACE"
  git clone "$GIT_REPO" .
  echo "Clone complete."
fi

exec "$@"
