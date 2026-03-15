# Agent State

CURRENT_PHASE: MVP
CURRENT_TASK: MVP-001
CURRENT_TASK_TITLE: Create project skeleton
STATUS: pending

LAST_COMPLETED_TASK:
LAST_COMPLETED_AT:

NEXT_TASK_HINT: MVP-002 Add config loader

RULES:
- execute only CURRENT_TASK
- do not start another task automatically unless CURRENT_TASK is fully complete
- after completion, mark CURRENT_TASK as done in the appropriate task file
- then move to the next unchecked task
- update this file
- append a short note to done_log.md
