# FarmSense, TaskAm and Logistics contract v1

The canonical fixtures are under `contracts/platform/v1`. External identifiers are opaque strings: `platform_user_id`, `farmsense_request_id`, `marketplace_request_id`, and `logistics_delivery_id`.

## Authentication

Service requests use `X-Platform-Service`, Unix `X-Platform-Timestamp`, unique `X-Platform-Nonce`, and `X-Platform-Signature`. The signature is `sha256=` plus hex HMAC-SHA256 of `METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + SHA256_HEX(RAW_BODY)`. Requests outside five minutes and replayed nonces are rejected. FarmSense and TaskAm have separate secrets. JWT, where retained by another service, is compatibility-only.

TaskAm request creation is exactly `POST /api/v1/internal/service-requests` and requires `Idempotency-Key`.

## Logistics URLs

- `POST /api/v1/deliveries/quotes`, requiring `Idempotency-Key`, accepts the quote request fixture and returns the quote response fixture.
- `POST /api/v1/deliveries`, requiring `Idempotency-Key`, accepts the booking request fixture and returns the booking response fixture.
- `GET /api/v1/deliveries/{logistics_delivery_id}` returns the booking/status DTO.

Legacy `/internal/v1/agricultural` URLs remain temporarily available for backward compatibility. They are not the canonical FarmSense integration.

## Lifecycle and events

The boundary lifecycle is `requested`, `quoted`, `confirmed`, `booked`, `provider_assigned`, `picked_up`, `in_transit`, `delivered`, `cancelled`, `failed`, `disputed`. Logistics retains its internal order statuses and maps them at the API/webhook boundary.

Status webhooks use the exact envelope fields `event_id`, `event_type`, `occurred_at`, `source`, the four shared IDs, and `data`. Signatures are `sha256=` plus hex HMAC-SHA256 of `TIMESTAMP + "." + RAW_BODY`; headers are `X-Logistics-Event-ID`, `X-Logistics-Timestamp`, and `X-Logistics-Signature`. Consumers must deduplicate by event ID.

## Driver evidence

The assigned driver/provider records pickup proof at `POST /api/v1/orders/{id}/proofs/pickup` and delivery proof at `/proofs/delivery`. Evidence must be a durable object URL. Delivery proof requires `recipient_name`.
