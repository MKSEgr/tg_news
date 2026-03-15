# MVP Spec

## Scope

The MVP must implement:

- project skeleton
- config loader
- structured logging
- graceful shutdown
- health endpoint
- postgres bootstrap
- redis bootstrap
- docker compose
- migrations
- domain models
- repository interfaces
- PostgreSQL repositories
- source/channel seed logic
- collector framework
- collectors:
  - rss
  - github
  - reddit
  - producthunt
- normalizer
- dedup
- trend scoring
- channel routing
- Yandex AI client
- content generator
- editorial guard
- scheduler
- Telegram publisher
- orchestration jobs
- admin HTTP API
- core tests

## Default publish mode

manual approve

## Constraints

- single Go application
- no vector DB
- no Kafka
- no ClickHouse
- no full web CMS in MVP
- no image generation in MVP
- no ML-heavy ranking in MVP

## Acceptance highlights

- system boots via Docker Compose
- health endpoint works
- new content is collected and stored
- duplicates are filtered
- items receive score and channel matches
- drafts are generated
- drafts can be approved
- approved drafts can be published to Telegram
