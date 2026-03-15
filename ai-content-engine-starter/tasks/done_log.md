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
