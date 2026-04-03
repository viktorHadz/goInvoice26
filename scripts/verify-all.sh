#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"$repo_root/scripts/check-frontend.sh"
"$repo_root/scripts/check-backend.sh"

echo "[verify] All frontend and backend checks passed."
