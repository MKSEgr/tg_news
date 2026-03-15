# Agent State

CURRENT_PHASE: MVP
CURRENT_TASK: MVP-009
CURRENT_TASK_TITLE: Add initial migrations
STATUS: pending

LAST_COMPLETED_TASK: MVP-008
LAST_COMPLETED_AT: 2026-03-15

NEXT_TASK_HINT: MVP-010 Implement domain models

RULES:
- execute only CURRENT_TASK
- do not start another task automatically unless CURRENT_TASK is fully complete
- after completion, mark CURRENT_TASK as done in the appropriate task file
- then move to the next unchecked task
- update this file
- append a short note to done_log.md
