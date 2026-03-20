# V2 Cleanup Plan (rollout-focused)

This is a constrained cleanup plan tied to rollout safety and architecture review outcomes.

## Critical now (before/at rollout)

1. **Runtime admin composition correctness**
   - Keep admin endpoints backed by a functional repository in runtime composition.
   - Validation: `/admin/drafts` operational checks in smoke tests.

2. **Feature flag governance**
   - Enforce global gate + per-feature toggles from a single config surface.
   - Validation: startup fails for invalid toggle combinations.

3. **Runbook-level rollback discipline**
   - For each enabled feature group, define immediate disable path and required telemetry capture.

## Medium-term (during V2 stabilization)

1. **Publishing lifecycle ownership clarity**
   - Document authoritative owner of status transitions (`pending` -> `approved` -> `posted`).

2. **Operational resilience improvements**
   - Add minimal retry/backoff policies for scheduler-triggered external IO operations.

3. **Discovery policy drift safeguards**
   - Track discovery acceptance/block reasons to detect score-threshold drift.

## Defer to V3 (explicitly not in V2 rollout changes)

1. **Large orchestration refactor**
   - Keep current orchestration structure; avoid high-risk rewrites during rollout.

2. **Repository interface decomposition**
   - Defer broad contract reshaping unless required by incident-driven needs.

3. **Deep schema partitioning**
   - Preserve current tables/semantics for V2; revisit scale partitioning in V3 planning.
