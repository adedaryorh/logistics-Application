# Logistics Platform

A Go microservice workspace for delivery and mobility workflows. The platform
separates identity, ordering, dispatch, location, payments, notifications, and
external access while sharing common infrastructure and service lifecycle code.

## Services

| Service | Responsibility | HTTP | gRPC |
| --- | --- | ---: | ---: |
| API gateway | Public API entrypoint and service routing | 8080 | — |
| Identity | Authentication, OAuth, users, and tokens | 8081 | 9081 |
| Logistics | Merchants, orders, assignments, and dispatch | 8082 | 9082 |
| Mobility | Driver location, routing, tracking, and geofencing | 8083 | 9083 |
| Payment | Transactions, wallets, refunds, and payment providers | 8084 | 9084 |
| Operations | Notifications, templates, flags, and background jobs | 8085 | 9085 |
| MCP server | Tool-facing platform integration | 8090 | — |

Every HTTP service exposes `/healthz` and `/metrics`. Stateful services use the
shared lifecycle in `pkg/servicehttp` for configuration, logging, telemetry,
middleware, HTTP/gRPC startup, background work, and graceful shutdown.

## Project layout

```text
services/        independently deployable applications
pkg/             small shared Go modules used across services
proto/           public and internal protobuf contracts
infra/           migrations, monitoring, and deployment configuration
scripts/         protobuf generation and verification
tests/           cross-service contract, integration, load, and E2E tests
mobile/          Expo/React Native customer application
```

Inside a stateful service, transport code belongs in `internal/http`, business
rules in `internal/service` (or `internal/application`), data types in
`internal/model`, and external persistence/provider details in repository or
infrastructure packages.

## Run locally

Requirements: Go 1.25.6 and Docker with Compose.

```bash
docker compose up --build
```

This starts PostgreSQL, Redis, Kafka, Schema Registry, Temporal, Jaeger,
Prometheus, Grafana, and all platform services. Useful local consoles are:

- Grafana: http://localhost:3000
- Temporal: http://localhost:8233
- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090

The credentials and keys in `docker-compose.yml` are development-only values.
Production deployments must supply secrets externally and enable the security
validation settings defined in `pkg/config`.

## Development commands

```bash
make build          # compile every service into bin/
make test           # run unit tests across all workspace modules
make proto-verify   # verify generated protobuf code is current
make docker-build   # build every service image
```

Integration suites require the Compose infrastructure and are intentionally
separate from the default unit test run. See the `test-*` targets in the
`Makefile` for contracts, workflows, migrations, load, Kafka, and chaos tests.

## Mobile app

The Expo client is in `mobile/`. Copy `mobile/.env.example` to `mobile/.env`,
set `EXPO_PUBLIC_API_URL` to the API gateway address, then run `npm start` from
the mobile directory. Use the computer's LAN address instead of `localhost`
when testing on a physical device.

## Adding a service

1. Create its module under `services/<name>` and add it to `go.work`.
2. Put its executable in `cmd/server` and domain code under `internal`.
3. For a stateful HTTP/gRPC service, implement `servicehttp.App` and pass its
   constructor through `servicehttp.Options.NewApp`.
4. Add the service to the `SERVICES` list, Compose configuration, protobuf
   contracts, migrations, and cross-service tests as applicable.
