#!/usr/bin/env bash

set -euo pipefail

PROTOC_GEN_GO_VERSION="${PROTOC_GEN_GO_VERSION:-v1.36.6}"
PROTOC_GEN_GO_GRPC_VERSION="${PROTOC_GEN_GO_GRPC_VERSION:-v1.5.1}"

go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"
go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"

echo "installed protoc Go plugins:"
echo "  protoc-gen-go ${PROTOC_GEN_GO_VERSION}"
echo "  protoc-gen-go-grpc ${PROTOC_GEN_GO_GRPC_VERSION}"
