# Database migrations

This directory contains SQL migrations for PostgreSQL.

Naming convention:
- `NNNNNN_description.up.sql` for apply scripts
- `NNNNNN_description.down.sql` for rollback scripts

Current bootstrap migration:
- `000001_initial_schema` creates core tables for channels, sources, source items, and drafts.
