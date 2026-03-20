# V2 Feature Flags Matrix

This matrix is the operational reference for V2 rollout.

## Global gate

| Flag | Default | Scope | Notes |
|---|---|---|---|
| `FEATURE_V2_ENABLED` | `false` | Global | Master gate. Must be `true` before any V2 sub-feature is enabled. |

## Feature flags

| Flag | Default | Classification | Depends on | Rollout notes |
|---|---|---|---|---|
| `FEATURE_ANALYTICS` | `false` | Operator/observability | `FEATURE_V2_ENABLED` | Enable first; non-invasive read-heavy feature. |
| `FEATURE_CONTENT_RULES` | `false` | Operator safety | `FEATURE_V2_ENABLED` | Enable before growth features to keep moderation controls. |
| `FEATURE_TOPIC_MEMORY` | `false` | Behavior-changing | `FEATURE_V2_ENABLED` | Enable after rules baseline; affects scoring/routing behavior. |
| `FEATURE_PERFORMANCE_FEEDBACK` | `false` | Behavior-changing | `FEATURE_V2_ENABLED` | Enable after analytics stability; influences routing/generation signals. |
| `FEATURE_AB_VARIANTS` | `false` | Experimental | `FEATURE_V2_ENABLED` | Enable after feedback semantics are stable. |
| `FEATURE_IMAGE_ENRICHMENT` | `false` | Behavior-changing | `FEATURE_V2_ENABLED` | Validate publish mode switching (text vs image) before broad rollout. |
| `FEATURE_SOURCE_DISCOVERY` | `false` | Experimental growth | `FEATURE_V2_ENABLED`, ideally rules+analytics | Enable late; verify duplicate suppression + quality gates. |
| `FEATURE_AUTO_REPOST` | `false` | Experimental growth | `FEATURE_V2_ENABLED`, feedback recommended | Enable late; monitor repost amplification effects. |
| `FEATURE_ADMIN_BOT` | `false` | Operator-only | `FEATURE_V2_ENABLED` | Enable after moderation lifecycle is stable. |
| `FEATURE_WEB_UI` | `false` | Operator-only | `FEATURE_V2_ENABLED` | Can be enabled early for visibility. |

## Enabled/disabled by default summary

- **Enabled by default**: none.
- **Disabled by default**: all V2 feature flags.

## Dependency rules

1. If any V2 feature flag is `true`, `FEATURE_V2_ENABLED` must be `true`.
2. Recommended soft dependencies for safe rollout:
   - `FEATURE_SOURCE_DISCOVERY` after `FEATURE_CONTENT_RULES` and `FEATURE_ANALYTICS`.
   - `FEATURE_AUTO_REPOST` after `FEATURE_PERFORMANCE_FEEDBACK`.
   - `FEATURE_AB_VARIANTS` after analytics/feedback validation.

## Operator-only vs experimental

- **Operator-only**: `FEATURE_ADMIN_BOT`, `FEATURE_WEB_UI`, `FEATURE_ANALYTICS`.
- **Experimental/growth**: `FEATURE_AB_VARIANTS`, `FEATURE_SOURCE_DISCOVERY`, `FEATURE_AUTO_REPOST`.
- **Core behavior modifiers**: `FEATURE_CONTENT_RULES`, `FEATURE_TOPIC_MEMORY`, `FEATURE_PERFORMANCE_FEEDBACK`, `FEATURE_IMAGE_ENRICHMENT`.

## Tiny central flag layer status

A minimal centralized flag layer already exists in `internal/platform/config` and enforces the global gate rule. No large refactor is required for V2 rollout safety.
