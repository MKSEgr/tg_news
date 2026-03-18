# Database migrations

This directory contains SQL migrations for PostgreSQL.

Naming convention:
- `NNNNNN_description.up.sql` for apply scripts
- `NNNNNN_description.down.sql` for rollback scripts

Current bootstrap migration:
- `000001_initial_schema` creates core tables for channels, sources, source items, and drafts.
- `000002_topic_memory` creates PostgreSQL-backed deterministic topic memory per channel.
- `000003_topic_memory_constraints` adds basic topic memory data quality checks.
- `000004_content_rules` creates PostgreSQL-backed blacklist/whitelist rules.
- `000005_content_rules_global_uniqueness` enforces uniqueness for global (channel_id NULL) rules.
- `000006_performance_feedback` creates PostgreSQL-backed deterministic performance feedback storage.


- `000007_ab_variants` — adds draft variants (`A`/`B`) for deterministic A/B generation and unique source/channel/variant storage.
- `000008_variant_attribution` — adds explicit A/B variant attribution to performance feedback rows for variant-level analytics.
- `000009_topic_memory_default_positive` — aligns `topic_memory.mention_count` default with positive-count constraint.
- `000013_content_assets` — creates minimal content assets storage for future asset-based generation.
- `000014_asset_relationships` — creates minimal explicit links between assets (`derived_from`, `followup_to`).

