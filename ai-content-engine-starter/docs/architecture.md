# Architecture

## High-level architecture

Single deployable Go application.

Core modules:
- collector
- normalizer
- dedup
- scorer
- router
- generator
- editorial guard
- scheduler
- publisher
- admin API

Infrastructure:
- PostgreSQL as primary storage
- Redis for locks and queues
- Docker Compose for local/dev

External integrations:
- Telegram Bot API
- Yandex AI Studio
- RSS
- GitHub API
- Reddit API
- Product Hunt API

## Principles

- keep it as one Go app
- prefer explicit code over clever abstractions
- PostgreSQL is the source of truth
- Redis is for locks, cooldowns, and queues
- modular architecture without microservice overhead
- feature flags for V2 features
- backward-compatible migrations when possible

## Pipeline

Sources
-> Collector
-> Normalizer
-> Dedup
-> Trend Scorer
-> Channel Router
-> Content Generator
-> Editorial Guard
-> Scheduler
-> Publisher
-> Telegram Channels

## Channels

### AI News
- model releases
- AI company news
- funding / acquisitions
- major research and product updates

### AI Tools
- product launches
- AI tools and SaaS
- open-source tooling
- tool cards and curated tool picks

### AI Workflows
- practical use cases
- automation workflows
- prompts
- business applications
- how-to content
