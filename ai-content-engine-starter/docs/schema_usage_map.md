# Schema Usage Map

## `channels`
- Purpose: channel catalog and routing targets.
- Writers:
  - `seed.Seeder` (default channels)
  - `postgres.ChannelRepository.Create`
- Readers:
  - `orchestration.PipelineJob` (channel list)
  - any channel CRUD consumers via repo
- Risks:
  - small table, low risk; slug uniqueness is key operational contract.

## `sources`
- Purpose: external source definitions (`kind`, `endpoint`, enabled flag).
- Writers:
  - `seed.Seeder`
  - `postgres.SourceRepository.Create`
- Readers:
  - `collector.Framework` (`ListEnabled`)
  - `orchestration.PipelineJob` (`ListEnabled`)
  - `sourcediscovery.Service` (`List` all sources for duplicate suppression)
- Risks:
  - endpoint quality/config errors can propagate to collector instability.

## `source_items`
- Purpose: normalized collected content items per source.
- Writers:
  - `collector.Framework` via `SourceItemRepository.Create`
- Readers:
  - `orchestration.PipelineJob` (`ListBySourceID`)
  - `dedup` checks (recent items)
- Risks:
  - potentially high growth; dedup and recent-item scans can become hot path.

## `drafts`
- Purpose: generated channel-targeted drafts with status and A/B variant.
- Writers:
  - `orchestration.PipelineJob` (`Create`)
  - admin moderation updates (`UpdateStatus`)
  - auto-repost promotion (`UpdateStatus`)
- Readers:
  - admin API list/get
  - analytics (`ListByStatus(posted)`)
  - feedback/auto-repost lookup flows
- Risks:
  - central table for many flows; status lifecycle consistency is critical.

## `topic_memory`
- Purpose: per-channel topic counters and recency for memory-aware ranking/routing.
- Writers:
  - `topicmemory.Service` (`UpsertMention`)
- Readers:
  - `orchestration.PipelineJob` optional memory reads
  - memory-aware scorer/router/guard paths
- Risks:
  - topic normalization quality affects signal reliability.

## `content_rules`
- Purpose: channel/global blacklist/whitelist rules.
- Writers:
  - `contentrules.Service.AddRule`
- Readers:
  - `contentrules.Service.Evaluate`
  - `orchestration.PipelineJob` rule gate
  - `sourcediscovery.DiscoverForChannel` candidate filtering
- Risks:
  - simplistic substring matching may over/under-filter depending on pattern quality.

## `performance_feedback`
- Purpose: engagement metrics + derived score per draft (with variant attribution).
- Writers:
  - `feedbackloop.Service` (`Upsert`)
- Readers:
  - `analytics.Service`
  - `orchestration` feedback-aware scorer/router/generator adjustments
  - `AutoRepostJob` candidate ranking
- Risks:
  - score semantics are central to multiple V2 behaviors; outliers/skew can bias routing and discovery gates.

## Cross-table usage hotspots

Most central write/read concentration:
1. `drafts`
2. `performance_feedback`
3. `source_items`
4. `sources`
5. `content_rules`

These tables carry the majority of orchestration and V2 control signals.
