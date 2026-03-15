# Agent State

CURRENT_PHASE: MVP
CURRENT_TASK: MVP-022
CURRENT_TASK_TITLE: Implement channel router
STATUS: pending

LAST_COMPLETED_TASK: MVP-021
LAST_COMPLETED_AT: 2026-03-15

NEXT_TASK_HINT: MVP-023 Implement Yandex AI client

RULES:
- execute only CURRENT_TASK
- do not start another task automatically unless CURRENT_TASK is fully complete
- after completion, mark CURRENT_TASK as done in the appropriate task file
- then move to the next unchecked task
- update this file
- append a short note to done_log.md
