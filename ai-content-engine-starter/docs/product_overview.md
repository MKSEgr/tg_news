# Product Overview

Project: AI content engine for 3 Telegram channels.

Channels:
1. AI News
2. AI Tools
3. AI Workflows

Goals:
- collect content from external sources
- normalize and deduplicate content
- score and route content into one or more channels
- generate draft posts with Yandex AI Studio
- queue and publish approved posts to Telegram
- start with manual approval, then evolve toward selective automation

Target audience:
- Russian-speaking and/or English-speaking audiences
- initial focus: AI / tools / automation / workflows content

Primary content sources:
- RSS feeds
- GitHub API
- Reddit API
- Product Hunt API

Non-goals for MVP:
- no Kafka
- no ClickHouse
- no vector database
- no complex ML ranking
- no full CMS
- no multi-service architecture

Success criteria for MVP:
- stable collection from sources
- working deduplication and routing
- draft generation for 3 channels
- manual approve workflow
- Telegram publishing works reliably
