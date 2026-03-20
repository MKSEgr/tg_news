# Start Here

## What is included

This starter pack contains:
- project docs
- MVP and V2 specs
- task lists
- agent state file
- reusable prompts for Cursor/Codex
- agent role notes

## Local run

For Docker-based local startup guidance (including the case where PostgreSQL and Redis are already running in other containers on your machine), see `docs/local_development.md`. The repository also keeps the runnable local stack definition in `docker-compose.yml`, so both files should remain committed alongside the project sources.

## Recommended workflow

1. Copy these files into the root of your repository.
2. Open the repository in Cursor or Codex.
3. Start with prompts/system_context.md.
4. Then use prompts/task_executor.md.
5. Let the agent implement only the current task from tasks/agent_state.md.
6. After each task, review the changes with prompts/reviewer.md.
7. Commit after each completed task.

## Suggested first run prompt

Use the repository task workflow.

Always start by reading:
- tasks/agent_state.md
- tasks/mvp_tasks.md
- tasks/v2_tasks.md
- prompts/coding_rules.md

Then execute only the current task.
When finished:
- mark it done
- move agent_state to the next unfinished task
- append a short log entry

Never skip ahead unless the task is blocked.
Never silently change the architecture.
