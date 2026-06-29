# Protobuf Workflow

This repository already contains the source `.proto` contracts under `/proto`.

## Required tools

Install these before generating Go stubs:

- `protoc`
- `protoc-gen-go`
- `protoc-gen-go-grpc`

Optional:

- `buf` if you want linting and generation through Buf instead of raw `protoc`

## Standard command

Use the repository make target:

```bash
make proto
```

That runs:

```bash
bash scripts/proto-generate.sh
```

The script:

- validates the required tools exist
- finds all `.proto` files under `proto/`
- generates Go and gRPC files with source-relative paths

## Install Go plugins

To install the Go protobuf generators with the pinned versions used by this repo:

```bash
bash scripts/proto-install-go-plugins.sh
```

## Tool check only

If you just want to confirm your machine is ready:

```bash
make proto-check-tools
```

## Lint contracts

If `buf` is installed:

```bash
make proto-lint
```

## Verify generated files

To regenerate stubs and fail if committed generated files are stale:

```bash
make proto-verify
```

## Buf option

Buf config is included:

- `buf.yaml`
- `buf.gen.yaml`

If you use Buf, the equivalent flow is:

```bash
buf lint
buf generate
```

CI now runs protobuf linting and generated-file verification automatically.

## Current backend note

The backend currently still uses the hand-maintained transport layer in `pkg/platformrpc` underneath the newer service-specific client packages. The protobuf generation workflow added here is the setup you will use when you are ready to switch fully to generated stubs.
