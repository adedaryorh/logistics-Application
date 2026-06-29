#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}"

bash scripts/proto-generate.sh

if ! git diff --exit-code -- '*.pb.go' '*.grpc.pb.go'; then
  echo "generated protobuf files are out of date" >&2
  exit 1
fi

echo "protobuf generated files are up to date"
