# Module Inventory (implementation-oriented)

## `internal/app`
- Responsibility: process lifecycle, config validation, HTTP route registration, graceful shutdown.
- Main deps: `platform/config`, `platform/logger`, `platform/postgres` (validation), `platform/redis` (validation), `admin`, `webui`.
- Depended on by: `cmd/app`.
- Coupling/breadth note: currently includes fallback admin repository wiring; operationally broad for a bootstrap module.

## `internal/domain`
- Responsibility: core entities + repository interfaces.
- Main deps: stdlib only.
- Depended on by: almost all business modules and Postgres adapters.
- Coupling/breadth note: central and stable; interface changes have large ripple effects.

## `internal/platform/postgres`
- Responsibility: concrete repository implementations and DB validation helpers.
- Main deps: `database/sql`, `internal/domain`.
- Depended on by: seed, orchestration wiring/tests, all repo-driven services when using Postgres.
- Coupling/breadth note: broad file with many repositories; high centrality.

## `internal/collector` (+ `collector/rss`, `collector/github`, `collector/reddit`, `collector/producthunt`)
- Responsibility: source ingestion from external providers and persistence through framework.
- Main deps: `internal/domain`, net/http/json/xml parsing.
- Depended on by: `orchestration.CollectorJob` (via runner abstraction).
- Coupling/breadth note: good separation by provider; framework is focused.

## `internal/normalizer`
- Responsibility: canonicalize source item URL/title/body text.
- Main deps: `internal/domain`.
- Depended on by: orchestration pipeline.
- Coupling/breadth note: focused.

## `internal/imageenrichment`
- Responsibility: deterministic image URL extraction from source items.
- Main deps: `internal/domain`.
- Depended on by: optional pipeline enricher.
- Coupling/breadth note: focused.

## `internal/dedup`
- Responsibility: duplicate detection using recent persisted source items.
- Main deps: `internal/domain`.
- Depended on by: pipeline.
- Coupling/breadth note: focused; DB read strategy can become hot path.

## `internal/scorer`
- Responsibility: trend scoring (base + optional memory/feedback adjustments).
- Main deps: `internal/domain`.
- Depended on by: pipeline.
- Coupling/breadth note: logic branch growth through optional interfaces.

## `internal/router`
- Responsibility: channel routing (base + optional memory/feedback routes).
- Main deps: `internal/domain`.
- Depended on by: pipeline.
- Coupling/breadth note: moderate complexity from optional variants.

## `internal/generator`
- Responsibility: draft generation using AI client, feedback-aware prompts, A/B variants.
- Main deps: `internal/domain` + AI client abstraction.
- Depended on by: pipeline.
- Coupling/breadth note: prompt building and variant logic in one service.

## `internal/editorial`
- Responsibility: deterministic editorial acceptance checks.
- Main deps: `internal/domain`.
- Depended on by: pipeline.
- Coupling/breadth note: focused.

## `internal/contentrules`
- Responsibility: create/evaluate blacklist/whitelist rules.
- Main deps: `internal/domain`.
- Depended on by: pipeline and discovery (V2-016 filtering).
- Coupling/breadth note: focused; simple string matching only.

## `internal/topicmemory`
- Responsibility: update/list top topics per channel.
- Main deps: `internal/domain`.
- Depended on by: pipeline optional memory-aware branches.
- Coupling/breadth note: focused.

## `internal/feedbackloop`
- Responsibility: validation + upsert of engagement metrics and score.
- Main deps: `internal/domain`.
- Depended on by: operational ingestion paths.
- Coupling/breadth note: focused.

## `internal/analytics`
- Responsibility: build per-channel summaries from posted drafts + feedback.
- Main deps: `internal/domain`.
- Depended on by: discovery integration (via local metrics adapter), reporting use cases.
- Coupling/breadth note: focused.

## `internal/sourcediscovery`
- Responsibility: derive deterministic candidate sources from observed URLs; optional channel-level analytics/rules gating.
- Main deps: `internal/domain` and local analytics/rules abstractions.
- Depended on by: currently standalone service (not fully wired into app jobs).
- Coupling/breadth note: focused; now avoids direct `analytics` package type coupling.

## `internal/orchestration`
- Responsibility: collector job, pipeline job, auto-repost job; central composition of most modules.
- Main deps: `internal/domain`, `internal/editorial`, multiple service interfaces.
- Depended on by: scheduler/application wiring (intended), tests.
- Coupling/breadth note: **most coupled and broadest module**.

## `internal/scheduler`
- Responsibility: periodic job execution loop.
- Main deps: stdlib context/time.
- Depended on by: app/runtime orchestration (intended).
- Coupling/breadth note: focused.

## `internal/publisher`
- Responsibility: Telegram publish path (`sendMessage`/`sendPhoto`).
- Main deps: `internal/domain`, net/http.
- Depended on by: publish runtime flow.
- Coupling/breadth note: focused.

## `internal/admin`
- Responsibility: draft moderation HTTP endpoints.
- Main deps: `internal/domain`.
- Depended on by: `internal/app` route registration.
- Coupling/breadth note: focused.

## `internal/adminbot`
- Responsibility: text command handling for moderation bot flows.
- Main deps: `internal/domain` abstractions.
- Depended on by: bot adapter wiring (intended).
- Coupling/breadth note: focused.

## `internal/webui`
- Responsibility: minimal HTML landing page.
- Main deps: stdlib net/http.
- Depended on by: app routes.
- Coupling/breadth note: focused.

## `internal/seed`
- Responsibility: bootstrap default channels/sources.
- Main deps: `internal/domain`.
- Depended on by: startup/bootstrap tasks (intended).
- Coupling/breadth note: focused.

## Platform helper modules
- `internal/platform/config`: env config loading + validation defaults.
- `internal/platform/logger`: JSON logger setup.
- `internal/platform/redis`: addr validation.
- `internal/platform/yandexai`: HTTP client adapter for generation backend.

These are mostly thin adapters and low coupling.
