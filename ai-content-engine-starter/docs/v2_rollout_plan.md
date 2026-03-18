# V2 Rollout Plan (based on architecture review)

This plan is explicitly derived from `docs/v2_architecture_review.md` and the current runtime/dataflow maps.

## Inputs from architecture review

Critical findings driving rollout order:
1. Runtime composition gap in app/admin wiring.
2. Missing centralized feature-flag control plane.
3. Cross-flow coupling via `drafts` + `performance_feedback` semantics.
4. Orchestration blast radius (defer major refactor; rollout with strict operational guardrails).

Medium findings influencing validation:
- publishing lifecycle ownership is implicit,
- resilience/retry behavior is basic,
- discovery gating is simple and can drift with scoring semantics.

## Default posture

- **Default = V2 disabled** (`FEATURE_V2_ENABLED=false`).
- Individual V2 features remain disabled until promoted phase-by-phase.
- Promote only one behavior-changing feature group at a time.
- Keep rollout reversible via flags (no schema rollback required for ordinary feature rollback).

## Rollout order of V2 modules

### Phase 0 — Platform safety baseline (must pass first)
Enabled:
- `FEATURE_V2_ENABLED=true` (control-plane only)
- `FEATURE_WEB_UI=true` (optional operator visibility)

Remain disabled:
- all behavior-modifying V2 pipeline features.

Operational checks before promotion:
- `/health` returns 200.
- admin moderation endpoints return successful list responses.
- config validation rejects invalid flag combinations.

Rollback:
- set `FEATURE_V2_ENABLED=false` and restart.

---

### Phase 1 — Observability-first
Enable first:
- `FEATURE_ANALYTICS=true`

Keep disabled:
- topic memory, rules, feedback-aware routing/generation, A/B variants, auto repost, source discovery.

Validation steps:
- analytics job/service produces deterministic per-channel summaries.
- verify `posted` drafts with and without feedback are handled correctly.
- confirm no routing/generation behavior change in this phase.

Operational checks:
- query latency and CPU for analytics paths.
- ensure no error spikes from feedback lookups.

Rollback:
- set `FEATURE_ANALYTICS=false`.

---

### Phase 2 — Content safety controls
Enable:
- `FEATURE_CONTENT_RULES=true`

Optional after stability window:
- `FEATURE_TOPIC_MEMORY=true`

Validation steps:
- rules: positive/negative pattern checks against known samples.
- memory: top-topic reads remain deterministic and channel-scoped.
- verify draft counts by status don’t collapse unexpectedly.

Operational checks:
- monitor blocked-vs-generated ratio.
- monitor moderation queue volume and rule false positives.

Rollback:
- disable the specific flag (`FEATURE_CONTENT_RULES` or `FEATURE_TOPIC_MEMORY`).

---

### Phase 3 — Feedback-influenced decisions
Enable:
- `FEATURE_PERFORMANCE_FEEDBACK=true`

Validation steps:
- feedback upserts remain idempotent per draft.
- score distributions are sane (no NaN/Inf; expected ranges).
- routing/generation changes are explainable from feedback data.

Operational checks:
- compare pre/post routing split and average draft quality metrics.
- watch for disproportionate amplification of single channels/topics.

Rollback:
- `FEATURE_PERFORMANCE_FEEDBACK=false`.

---

### Phase 4 — Experimentation and growth
Enable gradually (in order):
1. `FEATURE_AB_VARIANTS=true`
2. `FEATURE_IMAGE_ENRICHMENT=true`
3. `FEATURE_SOURCE_DISCOVERY=true`
4. `FEATURE_AUTO_REPOST=true`

Validation steps by feature:
- A/B: variant attribution and feedback linkage integrity.
- image enrichment: sendPhoto/sendMessage selection correctness.
- discovery: duplicate suppression against all sources, rules gate, analytics threshold behavior.
- auto repost: cooldown + min score + max-per-run respected.

Operational checks:
- draft volume growth and pending queue saturation.
- source growth quality (new discovered source false positives).
- repost rate vs original post rate.

Rollback:
- disable only the affected feature flag; preserve prior stable phases.

---

### Phase 5 — Operator surfaces
Enable:
- `FEATURE_ADMIN_BOT=true`
- keep `FEATURE_WEB_UI=true`

Validation steps:
- bot command allowlist behavior and moderation actions.
- operator workflows for list/approve/reject complete without manual DB intervention.

Operational checks:
- bot error rates, moderation action latency.

Rollback:
- disable `FEATURE_ADMIN_BOT`.

## Must-fix before production rollout

1. Runtime admin composition must remain functional (no always-unavailable admin repository path in active runtime).
2. Centralized feature-flag validation must gate partial V2 activation.
3. For every enabled phase, run canary + staged rollout with explicit rollback playbook.

## Rollback strategy (global)

If production instability appears:
1. Disable newest phase flag(s) first.
2. If instability persists, set `FEATURE_V2_ENABLED=false`.
3. Keep migrations in place; rollback behavior by flags, not destructive schema changes.
4. Preserve diagnostic artifacts (errors, latency, draft status deltas, feedback score distribution) before re-enable.
