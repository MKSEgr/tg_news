# Architecture Map (as implemented)

## System overview

This repository implements a **single Go application** with modular packages for content ingestion, processing, moderation, publishing, and V2 optimization features.

At runtime, the entrypoint (`cmd/app/main.go`) starts `internal/app.App`, which validates config (Postgres DSN + Redis addr), starts HTTP server routes, and exposes `/health`, `/admin/*`, and `/` web UI.

Key external dependencies used by modules:
- PostgreSQL (primary persistence via `internal/platform/postgres` repositories)
- Redis (validated in config path, not deeply integrated in orchestration yet)
- External source APIs (RSS, GitHub, Reddit, Product Hunt)
- Yandex AI API (text generation)
- Telegram Bot API (publishing)

## Main pipeline

Implemented end-to-end orchestration is centered in `internal/orchestration`:

1. Collection job:
   - `collector.Framework.RunOnce` gets enabled sources
   - dispatches to matching collector by source kind
   - persists `source_items`

2. Pipeline job (draft generation):
   - load enabled sources + channels + recent items
   - normalize -> optional image enrich -> dedup
   - scoring/routing (with optional memory/feedback-aware interfaces)
   - optional content rules gate per channel
   - generate single or A/B drafts
   - editorial guard (with optional memory-aware interface)
   - persist drafts with status pending/rejected

3. Auto-repost job:
   - load posted drafts and feedback
   - score/rank candidates
   - promote top drafts back to approved status for reposting

## Major V2 modules (implemented)

- **Topic memory** (`internal/topicmemory`): tracks per-channel topic mention counts.
- **Content rules** (`internal/contentrules`): blacklist/whitelist matching for channel text gates.
- **Feedback loop** (`internal/feedbackloop`): stores engagement metrics and computed score.
- **A/B variants** (generator + orchestration + schema): stores and handles variant A/B drafts.
- **Analytics** (`internal/analytics`): per-channel summaries from posted drafts + feedback.
- **Image enrichment** (`internal/imageenrichment`): image URL extraction from source items.
- **Source discovery** (`internal/sourcediscovery`): deterministic candidate source derivation, now optionally filtered by analytics/rules for a channel.
- **Admin bot + admin HTTP + web UI** (`internal/adminbot`, `internal/admin`, `internal/webui`): moderation and operational surfaces.

## Module interaction map

```text
Collectors -> source_items (DB)
source_items -> PipelineJob -> drafts (DB)

drafts (posted) + performance_feedback -> analytics

drafts -> feedbackloop.Upsert -> performance_feedback

topicmemory + contentrules + feedback + analytics -> influence pipeline/discovery decisions

publisher <- approved drafts (outside this repo wiring)
admin/api <- drafts moderation (list/approve/reject)
```

## Major architectural risk areas

1. **Runtime wiring gap in app bootstrap**
   - `internal/app` currently wires admin routes using a fallback in-memory-unavailable repository rather than real Postgres repos, so app HTTP surface is not yet fully connected to real storage.

2. **Orchestration complexity concentration**
   - `internal/orchestration/jobs.go` contains many optional interface branches (memory/feedback/variants/rules/image), making behavior hard to reason about and test regressions likely.

3. **Repository-centric coupling for all services**
   - Many modules depend directly on repository interfaces; changes to repository contracts cascade to many tests and stubs.

4. **No explicit feature-flag layer despite V2 principle**
   - V2 behavior is mostly enabled by dependency presence and type assertions rather than central feature-flag controls.

5. **Data lifecycle and moderation flow boundaries**
   - Draft statuses and repost promotion are deterministic, but ownership boundaries between scheduler/orchestration/admin/publisher are implicit and could drift without a stricter application service layer.
