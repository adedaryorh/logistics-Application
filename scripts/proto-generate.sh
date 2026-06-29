#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="${ROOT_DIR}/proto"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required tool: $1" >&2
    exit 1
  fi
}

need_cmd protoc
need_cmd protoc-gen-go
need_cmd protoc-gen-go-grpc

mapfile -t proto_files < <(find "${PROTO_DIR}" -name '*.proto' | sort)

if [[ ${#proto_files[@]} -eq 0 ]]; then
  echo "no proto files found under ${PROTO_DIR}" >&2
  exit 1
fi

cd "${ROOT_DIR}"

protoc \
  -I . \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  "${proto_files[@]}"

echo "generated protobuf stubs for ${#proto_files[@]} proto files"
