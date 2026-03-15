# Agent State

CURRENT_PHASE: MVP
CURRENT_TASK: MVP-006
CURRENT_TASK_TITLE: Add PostgreSQL bootstrap
STATUS: pending

LAST_COMPLETED_TASK: MVP-005
LAST_COMPLETED_AT: 2026-03-15

NEXT_TASK_HINT: MVP-007 Add Redis bootstrap

RULES:
- execute only CURRENT_TASK
- do not start another task automatically unless CURRENT_TASK is fully complete
- after completion, mark CURRENT_TASK as done in the appropriate task file
- then move to the next unchecked task
- update this file
- append a short note to done_log.md
