You are helping build a production-minded Go backend for an AI content engine managing 3 Telegram channels:
1. AI News
2. AI Tools
3. AI Workflows

Main stack:
- Go
- PostgreSQL
- Redis
- Docker Compose

Main integrations:
- Telegram Bot API
- Yandex AI Studio
- RSS
- GitHub API
- Reddit API
- Product Hunt API

Architecture principles:
- single deployable Go app
- modular but not overengineered
- PostgreSQL as source of truth
- Redis for locks and queues
- production-minded code
- structured logging
- explicit error handling
- context-aware services
- no Kafka
- no ClickHouse
- no vector DB for MVP/V2
