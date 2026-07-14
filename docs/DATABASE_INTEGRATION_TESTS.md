# Database-backed integration verification

Production sets `LOGISTICS_PERSISTENCE_MODE=sql`; development and unit tests default to the explicit in-memory store. SQL mode makes `/readyz` fail when PostgreSQL is unavailable.

With Docker Desktop running:

```sh
docker compose up -d postgres
MIGRATE_DATABASE_URL='postgres://postgres:postgres@localhost:5432/logistics?sslmode=disable' make migrate-up
TEST_INTEGRATION=1 TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/logistics?sslmode=disable' make test-migrations
```

Then start Logistics with the database variables used in `docker-compose.yml` and `LOGISTICS_PERSISTENCE_MODE=sql`. Configure `AGRICULTURAL_WEBHOOK_URL` and `AGRICULTURAL_WEBHOOK_SIGNING_SECRET` for durable delivery. Configure `EVIDENCE_UPLOAD_BASE_URL`, `EVIDENCE_PUBLIC_BASE_URL`, and `EVIDENCE_SIGNING_SECRET` for proof upload URLs.
