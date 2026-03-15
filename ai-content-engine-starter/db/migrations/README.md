# Database migrations

This directory contains SQL migrations for PostgreSQL.

Naming convention:
- `NNNNNN_description.up.sql` for apply scripts
- `NNNNNN_description.down.sql` for rollback scripts

Current bootstrap migration:
- `000001_initial_schema` creates core tables for channels, sources, source items, and drafts.
- `000002_topic_memory` creates PostgreSQL-backed deterministic topic memory per channel.
- `000003_topic_memory_constraints` adds basic topic memory data quality checks.

