# Agent State

CURRENT_PHASE: V2
CURRENT_TASK: V2-005
CURRENT_TASK_TITLE: Integrate rules into pipeline
STATUS: pending

LAST_COMPLETED_TASK: V2-004
LAST_COMPLETED_AT: 2026-03-15

NEXT_TASK_HINT: V2-006 Integrate feedback into scoring/generation/routing

RULES:
- execute only CURRENT_TASK
- do not start another task automatically unless CURRENT_TASK is fully complete
- after completion, mark CURRENT_TASK as done in the appropriate task file
- then move to the next unchecked task
- update this file
- append a short note to done_log.md
