#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v buf >/dev/null 2>&1; then
  echo "missing required tool: buf" >&2
  echo "install buf to run protobuf linting" >&2
  exit 1
fi

cd "${ROOT_DIR}"
buf lint

echo "protobuf lint passed"
