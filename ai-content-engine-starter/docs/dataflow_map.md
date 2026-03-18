# Dataflow Map (runtime behavior as implemented)

## 1) Source collection flow

```text
sources (enabled)
  -> collector.Framework.RunOnce
      -> collector by kind (rss/github/reddit/producthunt)
          -> []SourceItem
      -> source_items repository upsert/create
```

Notes:
- Framework selects collector by `source.kind`.
- Items are persisted with `(source_id, external_id)` uniqueness semantics in DB layer.

## 2) Content processing flow (pipeline core)

```text
PipelineJob.Run
  -> list enabled sources
  -> list channels
  -> list recent source_items per source
  -> normalize
  -> optional image enrichment
  -> dedup check
  -> score (base + optional memory/feedback adjustments)
  -> route channels (base + optional memory/feedback routing)
  -> optional content rules gate per target channel
```

Notes:
- Memory/rules/feedback are integrated via optional interfaces/dependencies.
- Existing drafts are preloaded by statuses to avoid duplicate draft creation.

## 3) Generation and publishing flow

```text
for each (item, channel)
  -> generator:
       - GenerateDraft OR
       - GenerateDraftWithFeedback OR
       - GenerateDraftVariants (A/B)
  -> editorial guard (optional memory-aware variant)
  -> drafts repository create (pending/rejected)

approved drafts
  -> publisher.Client.PublishDraft
      -> sendMessage (text) OR sendPhoto (image)
```

Notes:
- Publisher is implemented, but end-to-end runtime wiring from app to publishing loop is not fully assembled in `internal/app`.

## 4) Feedback / analytics flow

```text
engagement metrics input
  -> feedbackloop.Upsert
      -> performance_feedback table

analytics.BuildByChannel
  -> read posted drafts
  -> read feedback by draft
  -> compute per-channel averages + variant stats
```

Notes:
- Analytics ignores NaN/Inf scores.
- Variant-level averages (A/B) are derived from feedback records.

## 5) Topic memory / rules flow

```text
topicmemory.UpsertMention/ListTopByChannel
  -> topic_memory table

contentrules.AddRule/Evaluate
  -> content_rules table

PipelineJob
  -> optional topic memory read (for scorer/router/guard variants)
  -> optional rules evaluate per channel before draft generation
```

Notes:
- Rule evaluation is deterministic and channel-scoped with whitelist/blacklist behavior.

## 6) Source discovery flow (V2)

```text
Discover(items)
  -> list all existing sources (enabled + disabled)
  -> derive candidate endpoints from item URLs
  -> suppress duplicate host/endpoint candidates
  -> return sorted capped candidates

DiscoverForChannel(channelID, items)
  -> optional analytics gate (FeedbackDrafts > 0 && AvgScore < threshold => skip)
  -> Discover(items)
  -> optional content-rules filtering of candidates
```

Notes:
- Discovery currently returns candidate `domain.Source` objects and does not auto-persist by default.
- Integration uses local `ChannelMetrics` abstraction in `sourcediscovery` (no direct package coupling to `analytics` concrete type).
