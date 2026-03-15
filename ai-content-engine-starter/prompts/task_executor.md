Read the following files first:
- docs/product_overview.md
- docs/architecture.md
- spec/mvp_spec.md
- spec/v2_spec.md
- tasks/agent_state.md
- tasks/mvp_tasks.md
- tasks/v2_tasks.md
- prompts/coding_rules.md

Then do the following:
1. Find the CURRENT_TASK in tasks/agent_state.md
2. Implement only that task
3. Keep the implementation production-minded and minimal
4. Do not work on future tasks
5. After finishing:
   - mark the task as done in the correct task file
   - update tasks/agent_state.md to point to the next unfinished task
   - append a concise entry to tasks/done_log.md
6. Provide a short summary of what was changed and what remains

If the task is blocked, explain the blocker clearly and do not silently skip to another task.
