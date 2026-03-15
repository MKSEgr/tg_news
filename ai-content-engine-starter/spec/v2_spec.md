# V2 Spec

## V2 modules

- Telegram admin bot
- A/B variants
- auto repost of strong evergreen posts
- performance feedback loop
- topic memory
- blacklist/whitelist rules
- per-channel analytics
- basic web UI
- image enrichment
- automatic source discovery

## Rollout principles

- feature-flag driven
- backward-compatible DB migrations
- disabled by default until validated
- keep architecture inside the same Go application
- keep analytics in PostgreSQL for V2
- remain explainable and operationally simple

## V2 priorities

### First priority
- topic memory
- blacklist/whitelist rules
- performance feedback loop

### Second priority
- A/B variants
- variant attribution
- auto repost
- per-channel analytics

### Third priority
- Telegram admin bot
- basic web UI
- image enrichment
- source discovery

## Constraints

- no ClickHouse in V2
- no vector DB in V2
- no overbuilt experimentation framework
- no complex frontend stack unless clearly needed
