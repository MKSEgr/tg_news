# Done Log

## 2026-03-15
- Project scaffolding files prepared
  - docs/
  - spec/
  - tasks/
  - prompts/
  - agents/

- MVP-001 completed: created initial Go project skeleton (go.mod, cmd/app entrypoint, internal app package).
- MVP-002 completed: added environment-based config loader with defaults and validation for HTTP port.
- MVP-003 completed: added structured JSON logger with environment-based log level and app startup log.
- MVP-004 completed: added graceful shutdown handling for SIGINT/SIGTERM in app lifecycle.
- MVP-005 completed: added HTTP server startup/shutdown flow and /health endpoint returning status ok.
- MVP-006 completed: added PostgreSQL DSN config requirement and bootstrap validation package.
- MVP-007 completed: added Redis address config requirement and bootstrap validation package.
- MVP-008 completed: added Docker Compose for app, PostgreSQL, and Redis with healthchecks.
- MVP-009 completed: added initial PostgreSQL SQL migrations for channels, sources, source items, and drafts.
- MVP-010 completed: added core domain models and draft status constants aligned with initial schema.
- MVP-011 completed: added domain repository interfaces for channels, sources, source items, and drafts.
- MVP-012 completed: added PostgreSQL repository implementations for channels, sources, source items, and drafts.
- MVP-013 completed: added idempotent source/channel seed service with defaults and unit tests.
- MVP-014 completed: added collector framework orchestration with collector registry and persistence flow.
- MVP-015 completed: implemented RSS collector with HTTP fetch, XML parsing, and source item mapping.
- MVP-016 completed: implemented GitHub collector with API fetch, JSON parsing, and source item mapping.
- MVP-017 completed: implemented Reddit collector with listing fetch, JSON parsing, and source item mapping.
- MVP-018 completed: implemented Product Hunt collector with endpoint validation, optional endpoint-driven auth headers/token, JSON parsing, and source item mapping.
- MVP-019 completed: added normalizer service for canonical URL and text normalization with validation and unit tests.
