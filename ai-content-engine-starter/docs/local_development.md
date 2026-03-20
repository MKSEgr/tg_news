# Local development

## Repository files

The local-development setup in this starter is intentionally stored in version control:

- `docker-compose.yml` lives at the repository root.
- `docs/local_development.md` lives under `docs/`.

If you recreate or adjust these files locally, commit them like normal project files so other environments see the same startup instructions.

## Fastest startup when PostgreSQL and Redis are already running

If you already have PostgreSQL on `localhost:5432` and Redis on `localhost:6379` (for example from another Docker Compose project), the provided `docker-compose.yml` is already configured for that case:

```bash
docker compose up app
```

The `app` container reaches services published on your host through `host.docker.internal` and keeps the pipeline/publisher disabled by default, so only the HTTP/admin surfaces are started.

Default app environment in `docker-compose.yml`:

- `POSTGRES_DSN=postgres://postgres:postgres@host.docker.internal:5432/app?sslmode=disable`
- `REDIS_ADDR=host.docker.internal:6379`
- `ENABLE_PIPELINE=false`
- `ENABLE_PUBLISHER=false`
- `FEATURE_V2_ENABLED=true`
- `FEATURE_WEB_UI=true`

Open:

- `http://localhost:8080/health`
- `http://localhost:8080/health/runtime`
- `http://localhost:8080/`

## If you want this repository to start PostgreSQL and Redis too

Use the optional `infra` profile:

```bash
docker compose --profile infra up -d postgres redis
POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/app?sslmode=disable \
REDIS_ADDR=redis:6379 \
docker compose --profile infra up app
```

## Required migrations

The application validates PostgreSQL/Redis connectivity on startup, but it does not auto-apply SQL migrations. Apply the SQL files in `db/migrations/*.up.sql` before enabling the full pipeline against a fresh database.

Example with the optional local Postgres container from this repo:

```bash
for f in db/migrations/*up.sql; do
  docker compose --profile infra exec -T postgres psql -U postgres -d app < "$f"
done
```

## Enabling pipeline or publisher later

These toggles require extra credentials:

- `ENABLE_PIPELINE=true` requires `YANDEX_AI_API_KEY` and `YANDEX_AI_MODEL_URI`.
- `ENABLE_PUBLISHER=true` requires `TELEGRAM_BOT_TOKEN`.
