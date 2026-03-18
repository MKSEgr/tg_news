# V2-017 Final Architecture Review (code-aligned)

This review is based on the generated maps and direct inspection of runtime and module code.

Sources used:
- `docs/architecture_map.md`
- `docs/module_inventory.md`
- `docs/dataflow_map.md`
- `docs/schema_usage_map.md`
- key implementation packages under `internal/*`.

## Evaluation by requested areas

1. **Runtime wiring correctness**
   - Core HTTP lifecycle (`Run`, graceful shutdown, `/health`) is implemented correctly.
   - Major gap: `internal/app.routes()` wires admin with a fallback repository that always returns "unavailable" errors; no production storage wiring is present there.

2. **Orchestration complexity and responsibility boundaries**
   - `internal/orchestration` successfully coordinates collection, scoring/routing, generation, moderation, reposting.
   - Complexity is concentrated in one package with many optional interface branches (memory/rules/feedback/variants/enrichment), increasing interaction risk.

3. **Database semantic consistency (`drafts`, `performance_feedback`, `sources`)**
   - Semantics are mostly coherent and deterministic.
   - `drafts` status is the shared control plane for admin moderation, analytics inputs, and repost flows.
   - `performance_feedback` is consumed across analytics, scoring/routing adjustments, and repost ranking; a single scoring semantic drives many decisions.
   - `sources` semantics improved for discovery safety (duplicates suppressed against all sources, not only enabled).

4. **Repository interface coupling**
   - Domain interfaces are clear and testable.
   - Coupling is high: interface changes (even minimal) ripple across orchestration/services/postgres/tests.

5. **Feature isolation and feature flagging**
   - V2 features are mostly isolated by optional dependencies and interface detection.
   - There is no centralized feature flag/config layer enforcing rollout toggles at runtime.

6. **Publishing lifecycle ownership**
   - Publisher adapter is deterministic for text/image posting.
   - End-to-end ownership of state transitions into `posted` is distributed (orchestration/admin/publisher boundaries are implicit rather than explicitly centralized).

7. **Source discovery safety**
   - Good safeguards exist: normalized endpoint/host dedup, disabled-source suppression, analytics gating, and rule filtering.
   - Policy safety remains threshold-driven with a constant and can become sensitive if scoring semantics drift.

8. **Operational readiness (logging, health, retries)**
   - Health endpoint and structured logging exist.
   - Missing operational hardening in core loops/adapters: no explicit retry/backoff policies in scheduler/job execution path and external API adapters.

9. **Performance risks**
   - Pipeline and analytics frequently read broad sets (e.g., list by status with high/default limits), which can become expensive as tables grow.
   - Orchestration loads and branching in a single flow can become a scaling bottleneck without clearer stage boundaries.

10. **Documentation vs implementation alignment**
   - New architecture map docs are largely aligned with implemented modules and dataflow.
   - The main mismatch is operational completeness expectations versus current runtime wiring depth in `internal/app`.

---

## CRITICAL issues (must be addressed before V2 rollout)

1. **Runtime composition gap in `internal/app`**
   - Admin routes are exposed through an always-unavailable fallback draft repository, which means operationally "up" routes can still be non-functional for moderation workflows.

2. **Single-package orchestration blast radius**
   - `internal/orchestration` combines too many concerns in one coordination layer; optional dependency branches make regression risk high for rollout changes.

3. **No explicit rollout flag control plane for V2 features**
   - Behavior depends on dependency presence/type assertions rather than explicit, auditable feature flags, making controlled rollout and incident toggling harder.

4. **Cross-flow coupling through `drafts`/`performance_feedback` semantics**
   - Multiple behaviors depend on shared statuses/scores; changes in one flow can silently alter others (analytics, repost, routing/generation adjustments).

## MEDIUM issues (cleanup recommended)

1. **Publishing lifecycle ownership is implicit**
   - Clear ownership of transitions to `approved`/`posted` is not codified in a single service boundary.

2. **Repository interfaces are stable but broad in impact**
   - Minimal contract changes still require widespread test double updates.

3. **Operational resilience is basic**
   - Scheduler stops on first job error; retries/circuit-breaking are not first-class operational patterns yet.

4. **Postgres adapter density**
   - Many repository implementations in one file/package reduce locality and increase review burden.

5. **Discovery gating policy is simple but brittle to scoring drift**
   - Constant-threshold gating is deterministic, but not externally configurable for staged rollout tuning.

## LOW priority improvements

1. Document explicit draft status transition matrix with owner per transition.
2. Add a compact "optional-interface dependency map" for orchestration stages.
3. Document score semantics contract for `performance_feedback.Score` consumers.
4. Improve runbook docs for degraded mode (what remains functional when storage wiring is unavailable).
5. Split adapter docs/tests by repository domain for maintainability.

---

## Top 3 architectural strengths

1. **Strong modular decomposition with clear package purposes** (collector/normalizer/dedup/scorer/router/generator/etc.).
2. **Deterministic, test-heavy implementation style** with focused unit tests across modules.
3. **Pragmatic single-service architecture** (simple operational model, low distributed-systems overhead).

## Top 3 risks for V3 evolution

1. **Orchestration complexity growth** may outpace maintainability as more feature interactions are added.
2. **Semantic overload of shared control tables** (`drafts`, `performance_feedback`) may constrain independent evolution of analytics, moderation, and recommendation logic.
3. **Lack of centralized feature-flag governance** may make incremental rollout and rollback slower and riskier.
