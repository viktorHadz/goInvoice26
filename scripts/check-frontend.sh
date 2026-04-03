#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$repo_root/frontend"

echo "[frontend] Running type-check..."
npm run type-check

echo "[frontend] Running tests..."
npm test
