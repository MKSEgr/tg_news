# V2 Rollout Plan and Feature Flag Validation

## Goals
- Provide a single centralized feature-flag surface for V2 modules.
- Keep V2 features disabled by default.
- Prevent partial/accidental V2 activation through config validation.

## Centralized flags

The runtime now reads feature flags from environment in `internal/platform/config`:

- `FEATURE_V2_ENABLED`
- `FEATURE_TOPIC_MEMORY`
- `FEATURE_CONTENT_RULES`
- `FEATURE_PERFORMANCE_FEEDBACK`
- `FEATURE_AB_VARIANTS`
- `FEATURE_AUTO_REPOST`
- `FEATURE_ANALYTICS`
- `FEATURE_IMAGE_ENRICHMENT`
- `FEATURE_SOURCE_DISCOVERY`
- `FEATURE_ADMIN_BOT`
- `FEATURE_WEB_UI`

Validation rule:
- if any feature-specific flag is enabled, `FEATURE_V2_ENABLED` must be true.

## Rollout phases

1. **Phase 0 (default-safe)**
   - `FEATURE_V2_ENABLED=false`
   - all feature flags false

2. **Phase 1 (internal validation)**
   - `FEATURE_V2_ENABLED=true`
   - enable only observational features first (`FEATURE_ANALYTICS=true`)

3. **Phase 2 (controlled behavior changes)**
   - progressively enable routing/generation modifiers:
     - `FEATURE_TOPIC_MEMORY`
     - `FEATURE_CONTENT_RULES`
     - `FEATURE_PERFORMANCE_FEEDBACK`

4. **Phase 3 (distribution and growth)**
   - enable:
     - `FEATURE_IMAGE_ENRICHMENT`
     - `FEATURE_SOURCE_DISCOVERY`
     - `FEATURE_AUTO_REPOST`
     - `FEATURE_AB_VARIANTS`

5. **Phase 4 (operator surface)**
   - enable:
     - `FEATURE_ADMIN_BOT`
     - `FEATURE_WEB_UI`

## Runtime composition update

Admin runtime wiring no longer depends on an always-failing repository in app bootstrap. The app now wires an in-process draft repository so moderation endpoints remain operational by default while preserving compatibility with existing architecture.

## Validation checklist

- Config load fails when a feature-specific flag is enabled but `FEATURE_V2_ENABLED=false`.
- Config defaults leave all V2 features disabled.
- App routes expose working `/admin/drafts` endpoint under runtime wiring.
